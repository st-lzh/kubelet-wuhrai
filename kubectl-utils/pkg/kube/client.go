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

package kube

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client is a facade around the various kube interfaces
type Client struct {
	clientConfig    clientcmd.ClientConfig
	DyanmicClient   dynamic.Interface
	DiscoveryClient discovery.DiscoveryInterface
}

func NewClient(kubeconfig string) (*Client, error) {
	clientConfig, err := loadKubeconfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("building kubernetes API configuration: %w", err)
	}

	httpClient, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		return nil, fmt.Errorf("building http client for rest config: %w", err)
	}
	dynamicClient, err := dynamic.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return nil, fmt.Errorf("building dynamic client: %w", err)
	}
	discoveryClient, err := buildDiscoveryClient(restConfig, httpClient)
	if err != nil {
		return nil, err
	}
	return &Client{
		clientConfig:    clientConfig,
		DyanmicClient:   dynamicClient,
		DiscoveryClient: discoveryClient,
	}, nil
}

func loadKubeconfig(kubeconfigPath string) (clientcmd.ClientConfig, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfigPath != "" {
		rules.ExplicitPath = kubeconfigPath
	}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules,
		&clientcmd.ConfigOverrides{},
	)
	return clientConfig, nil
}

func (c *Client) DefaultNamespace() (string, error) {
	ns, _, err := c.clientConfig.Namespace()
	if err != nil {
		return "", fmt.Errorf("getting namespace from kubeconfig: %w", err)
	}
	namespace := ns
	if namespace == "" {
		namespace = "default"
	}
	return namespace, nil
}

// ForGVR returns a dynamic client for the specified GroupVersionResource and namespace
func (c *Client) ForGVR(gvr schema.GroupVersionResource, namespace string) dynamic.ResourceInterface {
	var client dynamic.ResourceInterface
	if namespace != "" {
		client = c.DyanmicClient.Resource(gvr).Namespace(namespace)
	} else {
		client = c.DyanmicClient.Resource(gvr)
	}
	return client
}
