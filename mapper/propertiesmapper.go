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

package mapper

import (
	"fmt"
	"strconv"
	"strings"
)

type PropertiesMapper struct {
	store
}

func NewPropertiesMapper() *PropertiesMapper {
	return &PropertiesMapper{store: newStore()}
}

func (m *PropertiesMapper) LoadFile(path string) error {
	return m.loadDocument(path, parseProperties)
}

func parseProperties(data []byte) (map[string]interface{}, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	out := make(map[string]interface{})
	var logical strings.Builder
	continued := false
	for _, line := range strings.Split(text, "\n") {
		if continued {
			line = trimLeftPropertySpace(line)
		}
		body, nextContinued := stripPropertyContinuation(line)
		logical.WriteString(body)
		if nextContinued {
			continued = true
			continue
		}
		if err := putProperty(out, logical.String()); err != nil {
			return nil, err
		}
		logical.Reset()
		continued = false
	}
	if continued {
		if err := putProperty(out, logical.String()); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func putProperty(out map[string]interface{}, line string) error {
	trimmed := trimLeftPropertySpace(line)
	if trimmed == "" || trimmed[0] == '#' || trimmed[0] == '!' {
		return nil
	}
	rawKey, rawValue := splitProperty(trimmed)
	key, err := unescapeProperties(rawKey)
	if err != nil {
		return err
	}
	value, err := unescapeProperties(rawValue)
	if err != nil {
		return err
	}
	out[key] = value
	return nil
}

func stripPropertyContinuation(line string) (string, bool) {
	slashes := 0
	for i := len(line) - 1; i >= 0 && line[i] == '\\'; i-- {
		slashes++
	}
	if slashes%2 == 1 {
		return line[:len(line)-1], true
	}
	return line, false
}

func splitProperty(line string) (string, string) {
	for i := 0; i < len(line); i++ {
		if line[i] == '\\' {
			i++
			continue
		}
		if line[i] == '=' || line[i] == ':' || isPropertySpace(line[i]) {
			rest := line[i:]
			if isPropertySpace(line[i]) {
				rest = trimLeftPropertySpace(rest)
				if len(rest) > 0 && (rest[0] == '=' || rest[0] == ':') {
					rest = trimLeftPropertySpace(rest[1:])
				}
			} else {
				rest = trimLeftPropertySpace(rest[1:])
			}
			return line[:i], rest
		}
	}
	return line, ""
}

func unescapeProperties(s string) (string, error) {
	if !strings.Contains(s, `\`) {
		return s, nil
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			continue
		}
		if i+1 >= len(s) {
			return "", fmt.Errorf("failed to read file: trailing escape in properties value")
		}
		i++
		switch s[i] {
		case 't':
			b.WriteByte('\t')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 'f':
			b.WriteByte('\f')
		case 'u':
			if i+4 >= len(s) {
				return "", fmt.Errorf("failed to read file: incomplete unicode escape")
			}
			hex := s[i+1 : i+5]
			n, err := strconv.ParseUint(hex, 16, 16)
			if err != nil {
				return "", fmt.Errorf("failed to read file: invalid unicode escape %q", `\u`+hex)
			}
			b.WriteRune(rune(n))
			i += 4
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String(), nil
}

func isPropertySpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\f'
}

func trimLeftPropertySpace(s string) string {
	i := 0
	for i < len(s) && isPropertySpace(s[i]) {
		i++
	}
	return s[i:]
}

var _ Mapper = (*PropertiesMapper)(nil)
