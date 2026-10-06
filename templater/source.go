// Copyright 2026 Richie Wood
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package templater

import (
	"fmt"
	"strings"
	"unicode"
)

type preparedSource struct {
	metadata     map[string]string
	includeFiles []string
	body         string
}

func prepareTemplate(source string) (preparedSource, error) {
	prepared := preparedSource{metadata: map[string]string{}}
	var body strings.Builder
	err := forEachLine(source, func(content string, newline bool) error {
		trimmed := strings.TrimLeftFunc(content, unicode.IsSpace)
		if strings.HasPrefix(trimmed, "#") {
			return recordDirective(strings.TrimLeft(trimmed, "#"), &prepared)
		}
		body.WriteString(content)
		if newline {
			body.WriteByte('\n')
		}
		return nil
	})
	if err != nil {
		return preparedSource{}, err
	}
	prepared.body = body.String()
	return prepared, nil
}

func prepareSource(source string) (string, error) {
	var body strings.Builder
	err := forEachLine(source, func(content string, newline bool) error {
		code := content
		if index := strings.IndexByte(content, '#'); index >= 0 {
			code = strings.TrimRightFunc(content[:index], unicode.IsSpace)
			if strings.TrimSpace(code) == "" {
				return nil
			}
		}
		body.WriteString(code)
		if newline {
			body.WriteByte('\n')
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return body.String(), nil
}

func recordDirective(comment string, prepared *preparedSource) error {
	directive := strings.TrimLeftFunc(comment, unicode.IsSpace)
	if !strings.HasPrefix(directive, "!") {
		return nil
	}
	directive = directive[1:]
	key, value, ok := strings.Cut(directive, ":")
	if !ok {
		return fmt.Errorf("metadata directive missing ':' in #!%s", directive)
	}
	key = strings.TrimSpace(key)
	if key == "" || strings.IndexFunc(key, unicode.IsSpace) >= 0 {
		return fmt.Errorf("invalid metadata key %q", key)
	}
	value = strings.TrimSpace(value)
	if key == "include-file" {
		prepared.includeFiles = append(prepared.includeFiles, value)
		return nil
	}
	prepared.metadata[key] = value
	return nil
}

func forEachLine(source string, fn func(content string, newline bool) error) error {
	rest := source
	for len(rest) > 0 {
		index := strings.IndexByte(rest, '\n')
		if index < 0 {
			return fn(rest, false)
		}
		content := strings.TrimSuffix(rest[:index], "\r")
		if err := fn(content, true); err != nil {
			return err
		}
		rest = rest[index+1:]
	}
	return nil
}
