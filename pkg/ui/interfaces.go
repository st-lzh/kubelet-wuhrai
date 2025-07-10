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

package ui

import "html/template"

type UI interface {
	// ClearScreen clears any output rendered to the screen
	ClearScreen()
}

type ComputedStyle struct {
	Foreground     ColorValue
	RenderMarkdown bool
}

type ColorValue string

const (
	ColorGreen ColorValue = "green"
	ColorWhite            = "white"
	ColorRed              = "red"
)

type StyleOption func(s *ComputedStyle)

func Foreground(color ColorValue) StyleOption {
	return func(s *ComputedStyle) {
		s.Foreground = color
	}
}

func RenderMarkdown() StyleOption {
	return func(s *ComputedStyle) {
		s.RenderMarkdown = true
	}
}

type CanFormatAsHTML interface {
	FormatAsHTML() template.HTML
}
