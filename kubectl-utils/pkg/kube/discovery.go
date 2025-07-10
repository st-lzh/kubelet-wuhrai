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
	"context"
	"fmt"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

func buildDiscoveryClient(restConfig *rest.Config, httpClient *http.Client) (discovery.DiscoveryInterface, error) {
	// TODO: share cache with kubectl?
	client, err := discovery.NewDiscoveryClientForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return nil, fmt.Errorf("building discovery client: %w", err)
	}
	return client, nil
}

func (c *Client) FindResource(ctx context.Context, name string) (*metav1.APIResource, error) {
	var matches []metav1.APIResource
	resourceLists, err := c.DiscoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("doing server discovery: %w", err)
	}
	for _, resourceList := range resourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			return nil, fmt.Errorf("parsing group version %q: %w", resourceList.GroupVersion, err)
		}
		for _, resource := range resourceList.APIResources {
			if resource.Kind == name {
				if resource.Group == "" {
					resource.Group = gv.Group
				}
				if resource.Version == "" {
					resource.Version = gv.Version
				}
				matches = append(matches, resource)
			}
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no match for resource %q", name)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("found multiple matches for resource %q", name)
	}
	resource := matches[0]
	return &resource, nil
}
