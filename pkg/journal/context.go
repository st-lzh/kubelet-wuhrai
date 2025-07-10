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
)

type contextKey string

const RecorderKey contextKey = "journal-recorder"

// RecorderFromContext extracts the recorder from the given context
func RecorderFromContext(ctx context.Context) Recorder {
	recorder, ok := ctx.Value(RecorderKey).(Recorder)
	if !ok {
		return &LogRecorder{}
	}
	return recorder
}

// ContextWithRecorder adds the recorder to the given context
func ContextWithRecorder(ctx context.Context, recorder Recorder) context.Context {
	return context.WithValue(ctx, RecorderKey, recorder)
}
