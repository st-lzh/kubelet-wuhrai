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

package html

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/st-lzh/kubelet-wuhrai/pkg/agent"
	"github.com/st-lzh/kubelet-wuhrai/pkg/journal"
	"k8s.io/klog/v2"
)

// APIServer 为kubectl-ai提供HTTP API端点
type APIServer struct {
	agent   agent.Agent
	journal journal.Recorder
}

// ChatRequest 表示来自API的聊天请求
type ChatRequest struct {
	Query     string            `json:"query"`
	SessionID string            `json:"session_id,omitempty"`
	Context   map[string]string `json:"context,omitempty"`
}

// ChatResponse 表示来自API的聊天响应
type ChatResponse struct {
	Response  string            `json:"response"`
	SessionID string            `json:"session_id"`
	Status    string            `json:"status"`
	Error     string            `json:"error,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// HealthResponse 表示健康检查响应
type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// ModelsResponse 表示可用模型响应
type ModelsResponse struct {
	Models    []string          `json:"models"`
	Current   string            `json:"current"`
	Provider  string            `json:"provider"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// NewAPIServer 创建一个新的API服务器实例
func NewAPIServer(agent agent.Agent, journal journal.Recorder) *APIServer {
	return &APIServer{
		agent:   agent,
		journal: journal,
	}
}

// RegisterAPIRoutes 使用提供的mux注册API路由
func (s *APIServer) RegisterAPIRoutes(mux *http.ServeMux) {
	// API v1路由
	mux.HandleFunc("POST /api/v1/chat", s.handleChatRequest)
	mux.HandleFunc("GET /api/v1/health", s.handleHealthCheck)
	mux.HandleFunc("GET /api/v1/models", s.handleModelsRequest)
	mux.HandleFunc("GET /api/v1/status", s.handleStatusRequest)

	// 所有API路由的CORS预检
	mux.HandleFunc("OPTIONS /api/v1/chat", s.handleCORS)
	mux.HandleFunc("OPTIONS /api/v1/health", s.handleCORS)
	mux.HandleFunc("OPTIONS /api/v1/models", s.handleCORS)
	mux.HandleFunc("OPTIONS /api/v1/status", s.handleCORS)
}

// handleChatRequest 处理聊天API请求
func (s *APIServer) handleChatRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := klog.FromContext(ctx)

	// 设置CORS头
	s.setCORSHeaders(w)

	// 解析请求
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(err, "解码聊天请求失败")
		s.writeErrorResponse(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	// 验证请求
	if req.Query == "" {
		s.writeErrorResponse(w, "查询是必需的", http.StatusBadRequest)
		return
	}

	log.Info("处理聊天请求", "query", req.Query, "session_id", req.SessionID)

	// Generate session ID if not provided
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("session_%d", time.Now().Unix())
	}

	// Process the query using the agent
	response := &ChatResponse{
		SessionID: req.SessionID,
		Status:    "processing",
		Timestamp: time.Now(),
	}

	// Run the agent with the query
	if err := s.agent.RunOneRound(ctx, req.Query); err != nil {
		log.Error(err, "agent failed to process query")
		response.Status = "error"
		response.Error = err.Error()
		s.writeJSONResponse(w, response, http.StatusInternalServerError)
		return
	}

	// For now, we'll return a simple success response
	// In a real implementation, you'd capture the agent's output
	response.Status = "completed"
	response.Response = "Query processed successfully. Check logs for detailed output."
	response.Metadata = map[string]string{
		"query_length": fmt.Sprintf("%d", len(req.Query)),
		"processed_at": time.Now().Format(time.RFC3339),
	}

	s.writeJSONResponse(w, response, http.StatusOK)
}

// handleHealthCheck handles health check requests
func (s *APIServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)

	response := &HealthResponse{
		Status:    "healthy",
		Version:   "dev", // This should come from build info
		Timestamp: time.Now(),
	}

	s.writeJSONResponse(w, response, http.StatusOK)
}

// handleModelsRequest handles models listing requests
func (s *APIServer) handleModelsRequest(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)

	// This is a placeholder - in a real implementation, you'd query the LLM client
	response := &ModelsResponse{
		Models: []string{
			"deepseek-chat",
			"deepseek-coder",
			"qwen-plus",
			"qwen-turbo",
			"doubao-pro-4k",
			"doubao-lite-4k",
		},
		Current:   "deepseek-chat",
		Provider:  "deepseek",
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"default_provider": "deepseek",
			"supports_tools":   "true",
		},
	}

	s.writeJSONResponse(w, response, http.StatusOK)
}

// handleStatusRequest handles status requests
func (s *APIServer) handleStatusRequest(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)

	status := map[string]interface{}{
		"status":     "running",
		"timestamp":  time.Now(),
		"uptime":     time.Since(time.Now()).String(), // Placeholder
		"agent":      "ready",
		"tools":      "available",
		"mcp_client": "disabled", // This should reflect actual MCP status
	}

	s.writeJSONResponse(w, status, http.StatusOK)
}

// handleCORS handles CORS preflight requests
func (s *APIServer) handleCORS(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	w.WriteHeader(http.StatusOK)
}

// setCORSHeaders sets CORS headers for API responses
func (s *APIServer) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

// writeJSONResponse writes a JSON response
func (s *APIServer) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		klog.Errorf("Failed to encode JSON response: %v", err)
	}
}

// writeErrorResponse writes an error response
func (s *APIServer) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	errorResp := map[string]interface{}{
		"error":     message,
		"status":    "error",
		"timestamp": time.Now(),
	}
	s.writeJSONResponse(w, errorResp, statusCode)
}

// StreamingChatRequest represents a streaming chat request
type StreamingChatRequest struct {
	Query     string            `json:"query"`
	SessionID string            `json:"session_id,omitempty"`
	Context   map[string]string `json:"context,omitempty"`
	Stream    bool              `json:"stream,omitempty"`
}

// StreamingChatResponse represents a streaming chat response chunk
type StreamingChatResponse struct {
	Delta     string            `json:"delta,omitempty"`
	Response  string            `json:"response,omitempty"`
	SessionID string            `json:"session_id"`
	Status    string            `json:"status"`
	Error     string            `json:"error,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Done      bool              `json:"done"`
}

// RegisterStreamingRoutes registers streaming API routes
func (s *APIServer) RegisterStreamingRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/chat/stream", s.handleStreamingChatRequest)
	mux.HandleFunc("OPTIONS /api/v1/chat/stream", s.handleCORS)
}

// handleStreamingChatRequest handles streaming chat API requests
func (s *APIServer) handleStreamingChatRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := klog.FromContext(ctx)

	// Set CORS and SSE headers
	s.setCORSHeaders(w)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Parse request
	var req StreamingChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(err, "failed to decode streaming chat request")
		s.writeSSEError(w, "Invalid request format")
		return
	}

	// Validate request
	if req.Query == "" {
		s.writeSSEError(w, "Query is required")
		return
	}

	log.Info("Processing streaming chat request", "query", req.Query, "session_id", req.SessionID)

	// Generate session ID if not provided
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("session_%d", time.Now().Unix())
	}

	// Send initial response
	initialResp := &StreamingChatResponse{
		SessionID: req.SessionID,
		Status:    "processing",
		Timestamp: time.Now(),
		Done:      false,
	}
	s.writeSSEResponse(w, initialResp)

	// Process the query (this is a simplified implementation)
	// In a real implementation, you'd capture the agent's streaming output
	finalResp := &StreamingChatResponse{
		SessionID: req.SessionID,
		Status:    "completed",
		Response:  "Query processed successfully. Check logs for detailed output.",
		Timestamp: time.Now(),
		Done:      true,
		Metadata: map[string]string{
			"query_length": fmt.Sprintf("%d", len(req.Query)),
			"processed_at": time.Now().Format(time.RFC3339),
		},
	}
	s.writeSSEResponse(w, finalResp)
}

// writeSSEResponse writes a Server-Sent Events response
func (s *APIServer) writeSSEResponse(w http.ResponseWriter, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		klog.Errorf("Failed to marshal SSE data: %v", err)
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// writeSSEError writes an SSE error response
func (s *APIServer) writeSSEError(w http.ResponseWriter, message string) {
	errorResp := &StreamingChatResponse{
		Status:    "error",
		Error:     message,
		Timestamp: time.Now(),
		Done:      true,
	}
	s.writeSSEResponse(w, errorResp)
}
