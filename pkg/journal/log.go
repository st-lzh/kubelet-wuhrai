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

package journal

import (
	"context"

	"k8s.io/klog/v2"
)

type LogRecorder struct {
}

func (r *LogRecorder) Write(ctx context.Context, event *Event) error {
	log := klog.FromContext(ctx)

	log.V(2).Info("Tracing event", "event", event)
	return nil
}

func (r *LogRecorder) Close() error {
	return nil
}
