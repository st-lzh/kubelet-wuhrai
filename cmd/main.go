// Copyright 2024 kubelet-wuhrai Contributors
//
// Licensed under the kubelet-wuhrai Custom License (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://github.com/st-lzh/kubelet-wuhrai/blob/main/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This software is based on kubectl-ai by Google Cloud Platform:
// https://github.com/GoogleCloudPlatform/kubectl-ai

package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/st-lzh/kubelet-wuhrai/gollm"
	"github.com/st-lzh/kubelet-wuhrai/pkg/agent"
	"github.com/st-lzh/kubelet-wuhrai/pkg/journal"
	"github.com/st-lzh/kubelet-wuhrai/pkg/mcp"
	"github.com/st-lzh/kubelet-wuhrai/pkg/tools"
	"github.com/st-lzh/kubelet-wuhrai/pkg/ui"
	"github.com/st-lzh/kubelet-wuhrai/pkg/ui/html"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

// Using the defaults from goreleaser as per https://goreleaser.com/cookbooks/using-main.version/
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func BuildRootCommand(opt *Options) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "kubelet-wuhrai",
		Short: "A CLI tool to interact with Kubernetes using natural language",
		Long:  "kubelet-wuhrai is a command-line tool that allows you to interact with your Kubernetes cluster using natural language queries. It leverages large language models to understand your intent and translate it into kubectl",
		Args:  cobra.MaximumNArgs(1), // Only one positional arg is allowed.
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunRootCommand(cmd.Context(), *opt, args)
		},
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of kubelet-wuhrai",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("version: %s\ncommit: %s\ndate: %s\n", version, commit, date)
			os.Exit(0)
		},
	})

	if err := opt.bindCLIFlags(rootCmd.Flags()); err != nil {
		return nil, err
	}
	return rootCmd, nil
}

type Options struct {
	ProviderID string `json:"llmProvider,omitempty"`
	ModelID    string `json:"model,omitempty"`
	// SkipPermissions is a flag to skip asking for confirmation before executing kubectl commands
	// that modifies resources in the cluster.
	SkipPermissions bool `json:"skipPermissions,omitempty"`
	// EnableToolUseShim is a flag to enable tool use shim.
	// TODO(droot): figure out a better way to discover if the model supports tool use
	// and set this automatically.
	EnableToolUseShim bool `json:"enableToolUseShim,omitempty"`
	// Quiet flag indicates if the agent should run in non-interactive mode.
	// It requires a query to be provided as a positional argument.
	Quiet     bool `json:"quiet,omitempty"`
	MCPServer bool `json:"mcpServer,omitempty"`
	MCPClient bool `json:"mcpClient,omitempty"`
	// ExternalTools enables discovery and exposure of external MCP tools (only works with --mcp-server)
	ExternalTools bool `json:"externalTools,omitempty"`
	MaxIterations int  `json:"maxIterations,omitempty"`

	// KubeConfigPath is the path to the kubeconfig file.
	// If not provided, the default kubeconfig path will be used.
	KubeConfigPath string `json:"kubeConfigPath,omitempty"`

	PromptTemplateFilePath string   `json:"promptTemplateFilePath,omitempty"`
	ExtraPromptPaths       []string `json:"extraPromptPaths,omitempty"`
	TracePath              string   `json:"tracePath,omitempty"`
	RemoveWorkDir          bool     `json:"removeWorkDir,omitempty"`
	ToolConfigPaths        []string `json:"toolConfigPaths,omitempty"`

	// UserInterface is the type of user interface to use.
	UserInterface UserInterface `json:"userInterface,omitempty"`
	// UIListenAddress is the address to listen for the HTML UI.
	UIListenAddress string `json:"uiListenAddress,omitempty"`

	// SkipVerifySSL is a flag to skip verifying the SSL certificate of the LLM provider.
	SkipVerifySSL bool `json:"skipVerifySSL,omitempty"`
}

type UserInterface string

const (
	UserInterfaceTerminal UserInterface = "terminal"
	UserInterfaceHTML     UserInterface = "html"
)

// Implement pflag.Value for UserInterface
func (u *UserInterface) Set(s string) error {
	switch s {
	case "terminal", "html":
		*u = UserInterface(s)
		return nil
	default:
		return fmt.Errorf("invalid user interface: %s", s)
	}
}

func (u *UserInterface) String() string {
	return string(*u)
}

func (u *UserInterface) Type() string {
	return "UserInterface"
}

var defaultToolConfigPaths = []string{
	filepath.Join("{CONFIG}", "kubelet-wuhrai", "tools.yaml"),
	filepath.Join("{HOME}", ".config", "kubelet-wuhrai", "tools.yaml"),
}

var defaultConfigPaths = []string{
	filepath.Join("{CONFIG}", "kubelet-wuhrai", "config.yaml"),
	filepath.Join("{HOME}", ".config", "kubelet-wuhrai", "config.yaml"),
}

func (o *Options) InitDefaults() {
	o.ProviderID = "deepseek"
	o.ModelID = "deepseek-chat"
	// by default, confirm before executing kubectl commands that modify resources in the cluster.
	o.SkipPermissions = false
	o.MCPServer = false
	o.MCPClient = false
	// by default, external tools are disabled (only works with --mcp-server)
	o.ExternalTools = false
	// DeepSeek models support tool use natively, so we don't need shim.
	o.EnableToolUseShim = false
	o.Quiet = false
	o.MCPServer = false
	o.MaxIterations = 20
	o.KubeConfigPath = ""
	o.PromptTemplateFilePath = ""
	o.ExtraPromptPaths = []string{}
	o.TracePath = filepath.Join(os.TempDir(), "kubelet-wuhrai-trace.txt")
	o.RemoveWorkDir = false
	o.ToolConfigPaths = defaultToolConfigPaths
	// Default to terminal UI
	o.UserInterface = UserInterfaceTerminal
	// Default UI listen address for HTML UI
	o.UIListenAddress = "localhost:8888"

	// Default to not skipping SSL verification
	o.SkipVerifySSL = false
}

func (o *Options) LoadConfiguration(b []byte) error {
	if err := yaml.Unmarshal(b, &o); err != nil {
		return fmt.Errorf("parsing configuration: %w", err)
	}
	return nil
}

func (o *Options) LoadConfigurationFile() error {
	configPaths := defaultConfigPaths
	for _, configPath := range configPaths {
		pathWithPlaceholdersExpanded := configPath

		if strings.Contains(pathWithPlaceholdersExpanded, "{CONFIG}") {
			configDir, err := os.UserConfigDir()
			if err != nil {
				return fmt.Errorf("getting user config directory (for config file path %q): %w", configPath, err)
			}
			pathWithPlaceholdersExpanded = strings.ReplaceAll(pathWithPlaceholdersExpanded, "{CONFIG}", configDir)
		}

		if strings.Contains(pathWithPlaceholdersExpanded, "{HOME}") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting user home directory (for config file path %q): %w", configPath, err)
			}
			pathWithPlaceholdersExpanded = strings.ReplaceAll(pathWithPlaceholdersExpanded, "{HOME}", homeDir)
		}

		configPath = filepath.Clean(pathWithPlaceholdersExpanded)
		configBytes, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				// ignore missing config files, they are optional
			} else {
				fmt.Fprintf(os.Stderr, "warning: could not load defaults from %q: %v\n", configPath, err)
			}
		}
		if len(configBytes) > 0 {
			if err := o.LoadConfiguration(configBytes); err != nil {
				fmt.Fprintf(os.Stderr, "warning: error loading configuration from %q: %v\n", configPath, err)
			}
		}
	}
	return nil
}

func main() {
	ctx := context.Background()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		klog.Flush()
		fmt.Fprintf(os.Stderr, "Received signal, shutting down... %s\n", sig)
		os.Exit(0)
	}()

	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// klog setup must happen before Cobra parses any flags

	// add commandline flags for logging
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	klogFlags.Set("logtostderr", "false")
	klogFlags.Set("log_file", filepath.Join(os.TempDir(), "kubelet-wuhrai.log"))

	defer klog.Flush()

	var opt Options

	opt.InitDefaults()

	// load YAML config values
	if err := opt.LoadConfigurationFile(); err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	rootCmd, err := BuildRootCommand(&opt)
	if err != nil {
		return err
	}

	// cobra has to know that we pass pass flags with flag lib, otherwise it creates conflict with flags.parse() method
	// We add just the klog flags we want, not all the klog flags (there are a lot, most of them are very niche)
	rootCmd.PersistentFlags().AddGoFlag(klogFlags.Lookup("v"))
	rootCmd.PersistentFlags().AddGoFlag(klogFlags.Lookup("alsologtostderr"))

	// do this early, before the third-party code logs anything.
	redirectStdLogToKlog()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return err
	}

	return nil
}

func (opt *Options) bindCLIFlags(f *pflag.FlagSet) error {
	f.IntVar(&opt.MaxIterations, "max-iterations", opt.MaxIterations, "代理在放弃之前尝试的最大迭代次数")
	f.StringVar(&opt.KubeConfigPath, "kubeconfig", opt.KubeConfigPath, "kubeconfig文件的路径")
	f.StringVar(&opt.PromptTemplateFilePath, "prompt-template-file-path", opt.PromptTemplateFilePath, "自定义提示模板文件的路径")
	f.StringArrayVar(&opt.ExtraPromptPaths, "extra-prompt-paths", opt.ExtraPromptPaths, "额外的提示模板路径")
	f.StringVar(&opt.TracePath, "trace-path", opt.TracePath, "跟踪文件的路径")
	f.BoolVar(&opt.RemoveWorkDir, "remove-workdir", opt.RemoveWorkDir, "执行后删除临时工作目录")

	f.StringVar(&opt.ProviderID, "llm-provider", opt.ProviderID, "语言模型提供商")
	f.StringVar(&opt.ModelID, "model", opt.ModelID, "语言模型，例如 deepseek-chat, deepseek-coder, qwen-plus, doubao-pro-4k")
	f.BoolVar(&opt.SkipPermissions, "skip-permissions", opt.SkipPermissions, "(危险) 跳过在执行修改资源的kubectl命令前的确认询问")
	f.BoolVar(&opt.MCPServer, "mcp-server", opt.MCPServer, "以MCP服务器模式运行")
	f.BoolVar(&opt.ExternalTools, "external-tools", opt.ExternalTools, "在MCP服务器模式下，发现并暴露外部MCP工具")
	f.StringArrayVar(&opt.ToolConfigPaths, "custom-tools-config", opt.ToolConfigPaths, "自定义工具配置文件或目录的路径")
	f.BoolVar(&opt.MCPClient, "mcp-client", opt.MCPClient, "启用MCP客户端模式以连接到外部MCP服务器")
	f.BoolVar(&opt.EnableToolUseShim, "enable-tool-use-shim", opt.EnableToolUseShim, "启用工具使用垫片")
	f.BoolVar(&opt.Quiet, "quiet", opt.Quiet, "以非交互模式运行，需要提供查询作为位置参数")

	f.Var(&opt.UserInterface, "user-interface", "要使用的用户界面模式。支持的值：terminal, html")
	f.StringVar(&opt.UIListenAddress, "ui-listen-address", opt.UIListenAddress, "HTML UI监听的地址")
	f.BoolVar(&opt.SkipVerifySSL, "skip-verify-ssl", opt.SkipVerifySSL, "跳过验证LLM提供商的SSL证书")

	return nil
}

func RunRootCommand(ctx context.Context, opt Options, args []string) error {
	var err error // Declare err once for the whole function

	// 验证标志组合
	if opt.ExternalTools && !opt.MCPServer {
		return fmt.Errorf("--external-tools只能与--mcp-server一起使用")
	}

	// 按优先级解析kubeconfig路径：标志/环境变量 > KUBECONFIG > 默认路径
	if err = resolveKubeConfigPath(&opt); err != nil {
		return fmt.Errorf("解析kubeconfig路径失败: %w", err)
	}

	if opt.MCPServer {
		if err = startMCPServer(ctx, opt); err != nil {
			return fmt.Errorf("启动MCP服务器失败: %w", err)
		}
		return nil // MCP server mode blocks, so we return here
	}

	if err := handleCustomTools(opt.ToolConfigPaths); err != nil {
		return fmt.Errorf("failed to process custom tools: %w", err)
	}

	// Initialize MCP client if requested
	var mcpManager *mcp.Manager
	if opt.MCPClient {
		var err error
		mcpManager, err = InitializeMCPClient()
		if err != nil {
			klog.Errorf("Failed to initialize MCP client: %v", err)
			os.Exit(1) // Fail fast instead of continuing with degraded functionality
		} else {
			klog.V(1).Info("MCP client initialization completed successfully")
		}
	}

	// After reading stdin, it is consumed
	var hasInputData bool
	hasInputData, err = hasStdInData()
	if err != nil {
		return fmt.Errorf("failed to check if stdin has data: %w", err)
	}

	// Handles positional args or stdin
	var queryFromCmd string
	queryFromCmd, err = resolveQueryInput(hasInputData, args)
	if err != nil {
		return fmt.Errorf("failed to resolve query input %w", err)
	}

	klog.Info("Application started", "pid", os.Getpid())

	var llmClient gollm.Client
	if opt.SkipVerifySSL {
		llmClient, err = gollm.NewClient(ctx, opt.ProviderID, gollm.WithSkipVerifySSL())
	} else {
		llmClient, err = gollm.NewClient(ctx, opt.ProviderID)
	}
	if err != nil {
		return fmt.Errorf("creating llm client: %w", err)
	}
	defer llmClient.Close()

	var recorder journal.Recorder
	if opt.TracePath != "" {
		var fileRecorder journal.Recorder
		fileRecorder, err = journal.NewFileRecorder(opt.TracePath)
		if err != nil {
			return fmt.Errorf("creating trace recorder: %w", err)
		}
		defer fileRecorder.Close()
		recorder = fileRecorder
	} else {
		// Ensure we always have a recorder, to avoid nil checks
		recorder = &journal.LogRecorder{}
		defer recorder.Close()
	}

	doc := ui.NewDocument()

	var userInterface ui.UI
	switch opt.UserInterface {
	case UserInterfaceTerminal:
		// since stdin is already consumed, we use TTY for taking input from user
		useTTYForInput := hasInputData

		var u ui.UI
		u, err = ui.NewTerminalUI(doc, recorder, useTTYForInput)
		if err != nil {
			return err
		}
		userInterface = u

	case UserInterfaceHTML:
		var u ui.UI
		u, err = html.NewHTMLUserInterface(doc, opt.UIListenAddress, recorder)
		if err != nil {
			return err
		}
		// Only run server if the UI is actually an HTML UI
		if htmlUI, ok := u.(*html.HTMLUserInterface); ok {
			go func() {
				if err := htmlUI.RunServer(ctx); err != nil {
					klog.Fatalf("error running http server: %v", err)
				}
			}()
		}
		userInterface = u

	default:
		return fmt.Errorf("user-interface mode %q is not known", opt.UserInterface)
	}

	conversation := &agent.Conversation{
		Model:              opt.ModelID,
		Kubeconfig:         opt.KubeConfigPath,
		LLM:                llmClient,
		MaxIterations:      opt.MaxIterations,
		PromptTemplateFile: opt.PromptTemplateFilePath,
		ExtraPromptPaths:   opt.ExtraPromptPaths,
		Tools:              tools.Default(),
		Recorder:           recorder,
		RemoveWorkDir:      opt.RemoveWorkDir,
		SkipPermissions:    opt.SkipPermissions,
		EnableToolUseShim:  opt.EnableToolUseShim,
		MCPClientEnabled:   opt.MCPClient,
	}

	err = conversation.Init(ctx, doc)
	if err != nil {
		return fmt.Errorf("starting conversation: %w", err)
	}
	defer conversation.Close()

	chatSession := session{
		model:        opt.ModelID,
		doc:          doc,
		ui:           userInterface,
		conversation: conversation,
		LLM:          llmClient,
		mcpManager:   mcpManager,
	}

	// Prepare MCP server status blocks only when MCP client is enabled
	var mcpBlocks []ui.Block
	if opt.MCPClient {
		if blocks, err := GetMCPServerStatusWithClientMode(opt.MCPClient, mcpManager); err == nil && len(blocks) > 0 {
			header := ui.NewAgentTextBlock().WithText("\nMCP Server Status:")
			mcpBlocks = append(mcpBlocks, header)
			mcpBlocks = append(mcpBlocks, blocks...)
			// Log MCP server status to log file
			klog.Info("MCP server status retrieved successfully for REPL startup")
		} else if err != nil {
			klog.Warningf("Failed to retrieve MCP server status for REPL startup: %v", err)
		}
	}

	if opt.Quiet {
		if queryFromCmd == "" {
			return fmt.Errorf("quiet mode requires a query to be provided as a positional argument")
		}
		return chatSession.answerQuery(ctx, queryFromCmd)
	}

	return chatSession.repl(ctx, queryFromCmd, mcpBlocks)
}

func handleCustomTools(toolConfigPaths []string) error {
	// resolve tool config paths, and then load and register custom tools from config files and dirs
	for _, path := range toolConfigPaths {
		pathWithPlaceholdersExpanded := path

		if strings.Contains(pathWithPlaceholdersExpanded, "{CONFIG}") {
			configDir, err := os.UserConfigDir()
			if err != nil {
				klog.Warningf("Failed to get user config directory for tools path %q: %v", path, err)
				continue
			}
			pathWithPlaceholdersExpanded = strings.ReplaceAll(pathWithPlaceholdersExpanded, "{CONFIG}", configDir)
		}

		if strings.Contains(pathWithPlaceholdersExpanded, "{HOME}") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				klog.Warningf("Failed to get user home directory for tools path %q: %v", path, err)
				continue
			}
			pathWithPlaceholdersExpanded = strings.ReplaceAll(pathWithPlaceholdersExpanded, "{HOME}", homeDir)
		}

		cleanedPath := filepath.Clean(pathWithPlaceholdersExpanded)

		klog.Infof("Attempting to load custom tools from processed path: %q (original value from config: %q)", cleanedPath, path)

		if err := tools.LoadAndRegisterCustomTools(cleanedPath); err != nil {
			if errors.Is(err, os.ErrNotExist) && !slices.Contains(defaultToolConfigPaths, path) {
				// user specified a directory that does not exist, we must error out
				return fmt.Errorf("custom tools directory not found (original value: %q, processed path: %q)", path, cleanedPath)
			} else {
				klog.Warningf("Failed to load or register custom tools (original value: %q, processed path: %q): %v", path, cleanedPath, err)
			}
		}
	}
	return nil
}

// session represents the user chat session (interactive/non-interactive both)
type session struct {
	model           string
	ui              ui.UI
	doc             *ui.Document
	conversation    *agent.Conversation
	availableModels []string
	LLM             gollm.Client
	mcpManager      *mcp.Manager
}

// repl is a read-eval-print loop for the chat session.
func (s *session) repl(ctx context.Context, initialQuery string, initialBlocks []ui.Block) error {
	for _, block := range initialBlocks {
		s.doc.AddBlock(block)
	}
	query := initialQuery
	if query == "" {
		s.doc.AddBlock(ui.NewAgentTextBlock().WithText("Hey there, what can I help you with today?"))
	}
	for {
		if query == "" {
			input := ui.NewInputTextBlock()
			input.SetEditable(true)
			s.doc.AddBlock(input)

			userInput, err := input.Observable().Wait()
			if err != nil {
				if err == io.EOF {
					// Use hit control-D, or was piping and we reached the end of stdin.
					// Not a "big" problem
					return nil
				}
				return fmt.Errorf("reading input: %w", err)
			}
			query = strings.TrimSpace(userInput)
		}

		switch {
		case query == "":
			continue
		case query == "reset":
			err := s.conversation.Init(ctx, s.doc)
			if err != nil {
				return err
			}
		case query == "clear":
			s.ui.ClearScreen()
		case query == "exit" || query == "quit":
			// s.ui.RenderOutput(ctx, "Alright...bye.\n")
			return nil
		default:
			if err := s.answerQuery(ctx, query); err != nil {
				errorBlock := &ui.ErrorBlock{}
				errorBlock.SetText(fmt.Sprintf("Error: %v\n", err))
				s.doc.AddBlock(errorBlock)
			}
		}
		// Reset query to empty string so that we prompt for input again
		query = ""
	}
}

func (s *session) listModels(ctx context.Context) ([]string, error) {
	if s.availableModels == nil {
		modelNames, err := s.LLM.ListModels(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing models: %w", err)
		}
		s.availableModels = modelNames
	}
	return s.availableModels, nil
}

func (s *session) answerQuery(ctx context.Context, query string) error {
	switch {
	case query == "model":
		infoBlock := &ui.AgentTextBlock{}
		infoBlock.AppendText(fmt.Sprintf("Current model is `%s`\n", s.model))
		s.doc.AddBlock(infoBlock)

	case query == "version":
		infoBlock := &ui.AgentTextBlock{}
		infoBlock.AppendText(fmt.Sprintf("Version: `%s`\n", version))
		s.doc.AddBlock(infoBlock)

	case query == "models":
		models, err := s.listModels(ctx)
		if err != nil {
			return fmt.Errorf("listing models: %w", err)
		}
		infoBlock := &ui.AgentTextBlock{}
		infoBlock.AppendText("\n  Available models:\n")
		infoBlock.AppendText(strings.Join(models, "\n"))
		s.doc.AddBlock(infoBlock)

	case query == "tools":
		if s.conversation == nil {
			return fmt.Errorf("listing tols: conversation is not initialized")
		}
		infoBlock := &ui.AgentTextBlock{}
		infoBlock.AppendText("\n  Available tools:\n")
		infoBlock.AppendText(strings.Join(s.conversation.Tools.Names(), "\n"))
		s.doc.AddBlock(infoBlock)

	default:
		return s.conversation.RunOneRound(ctx, query)
	}
	return nil
}

// Redirect standard log output to our custom klog writer
// This is primarily to suppress warning messages from
// genai library https://github.com/googleapis/go-genai/blob/6ac4afc0168762dc3b7a4d940fc463cc1854f366/types.go#L1633
func redirectStdLogToKlog() {
	log.SetOutput(klogWriter{})

	// Disable standard log's prefixes (date, time, file info)
	// because klog will add its own more detailed prefix.
	log.SetFlags(0)
}

// Define a custom writer that forwards messages to klog.Warning
type klogWriter struct{}

// Implement the io.Writer interface
func (writer klogWriter) Write(data []byte) (n int, err error) {
	// We trim the trailing newline because klog adds its own.
	message := string(bytes.TrimSuffix(data, []byte("\n")))
	klog.Warning(message)
	return len(data), nil
}

func hasStdInData() (bool, error) {
	hasData := false

	stat, err := os.Stdin.Stat()
	if err != nil {
		return hasData, fmt.Errorf("checking stdin: %w", err)
	}
	hasData = (stat.Mode() & os.ModeCharDevice) == 0

	return hasData, nil
}

// resolveQueryInput determines the query input from positional args and/or stdin.
// It supports:
// - 1 positional arg only -> kubelet-wuhrai "get pods"
// - stdin only -> echo "get pods" | kubelet-wuhrai
// - 1 positional arg + stdin (combined) -> kubelet-wuhrai get <<< "pods" or kubelet-wuhrai "get" <<< "pods"
// As default no positional arg nor stdin
func resolveQueryInput(hasStdInData bool, args []string) (string, error) {
	switch {
	case len(args) == 1 && !hasStdInData:
		// Use argument directly
		return args[0], nil

	case len(args) == 1 && hasStdInData:
		// Combine arg + stdin
		var b strings.Builder
		b.WriteString(args[0])
		b.WriteString("\n")

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			b.WriteString(scanner.Text())
			b.WriteString("\n")
		}
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		query := strings.TrimSpace(b.String())
		if query == "" {
			return "", fmt.Errorf("no query provided from stdin")
		}
		return query, nil

	case len(args) == 0 && hasStdInData:
		// Read stdin only
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		query := strings.TrimSpace(string(b))
		if query == "" {
			return "", fmt.Errorf("no query provided from stdin")
		}
		return query, nil

	default:
		// Case: No input at all — return empty string, no error
		return "", nil
	}
}

func resolveKubeConfigPath(opt *Options) error {
	switch {
	case opt.KubeConfigPath != "":
		// Already set from flag or viper env
	case os.Getenv("KUBECONFIG") != "":
		opt.KubeConfigPath = os.Getenv("KUBECONFIG")
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		opt.KubeConfigPath = filepath.Join(home, ".kube", "config")
	}

	// We resolve the kubeconfig path to an absolute path, so we can run kubectl from any working directory.
	if opt.KubeConfigPath != "" {
		p, err := filepath.Abs(opt.KubeConfigPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for kubeconfig file %q: %w", opt.KubeConfigPath, err)
		}
		opt.KubeConfigPath = p
	}

	return nil
}

func startMCPServer(ctx context.Context, opt Options) error {
	workDir := filepath.Join(os.TempDir(), "kubelet-wuhrai-mcp")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return fmt.Errorf("error creating work directory: %w", err)
	}
	mcpServer, err := newKubectlMCPServer(ctx, opt.KubeConfigPath, tools.Default(), workDir, opt.ExternalTools)
	if err != nil {
		return fmt.Errorf("creating mcp server: %w", err)
	}
	return mcpServer.Serve(ctx)
}
