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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"sigs.k8s.io/yaml"
)

// ParseEventsFromFile will read the events from the given file path
func ParseEventsFromFile(p string) ([]*Event, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("opening file %q: %w", p, err)
	}
	defer f.Close()

	return ParseEvents(f)
}

// ParseEvents will read the events from the reader
func ParseEvents(r io.Reader) ([]*Event, error) {
	var events []*Event

	scanner := bufio.NewScanner(r)
	scanner.Split(splitYAML)
	for scanner.Scan() {
		b := scanner.Bytes()

		event := &Event{}

		if err := yaml.Unmarshal(b, &event); err != nil {
			return nil, fmt.Errorf("parsing yaml: %w", err)
		}

		if event != nil {
			events = append(events, event)
		}
	}

	return events, nil
}

var yamlSep = []byte("\n---\n")

// splitYAML is a split function for a Scanner that returns each object in a yaml multi-object doc.
func splitYAML(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, yamlSep); i >= 0 {
		// We have a full object.
		return i + len(yamlSep), data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated object. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
