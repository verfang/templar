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
)

type includeFile struct {
	path      string
	alias     string
	delimiter *string
	columns   []string
}

func parseInclude(value string) (includeFile, error) {
	parts, err := splitFields(value)
	if err != nil {
		return includeFile{}, err
	}
	path := ""
	if len(parts) > 0 {
		path = strings.TrimSpace(parts[0])
	}
	if path == "" || strings.Contains(path, ":") {
		return includeFile{}, fmt.Errorf("include-file is missing a filename in %s", value)
	}
	var spec includeFile
	spec.path = path
	var alias string
	for _, part := range parts[1:] {
		key, field, ok := strings.Cut(part, ":")
		if !ok {
			return includeFile{}, fmt.Errorf("include-file field %s is missing ':'", strings.TrimSpace(part))
		}
		switch strings.TrimSpace(key) {
		case "as":
			alias = strings.TrimSpace(field)
		case "delimiter":
			decoded, err := unquote(strings.TrimSpace(field))
			if err != nil {
				return includeFile{}, err
			}
			spec.delimiter = &decoded
		case "columns":
			columns, err := parseColumns(strings.TrimSpace(field))
			if err != nil {
				return includeFile{}, err
			}
			spec.columns = columns
		default:
			return includeFile{}, fmt.Errorf("include-file field %s is not a filename, alias, delimiter, or column list", strings.TrimSpace(key))
		}
	}
	if alias == "" {
		return includeFile{}, fmt.Errorf("include-file %s is missing as", path)
	}
	if !validAlias(alias) {
		return includeFile{}, fmt.Errorf("include alias %s is not a name", alias)
	}
	spec.alias = alias
	return spec, nil
}

func splitFields(value string) ([]string, error) {
	var parts []string
	start := 0
	depth := 0
	var quote byte
	escaped := false
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if ch == quote {
				quote = 0
			}
			continue
		}
		switch ch {
		case '"', '\'':
			quote = ch
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				parts = append(parts, value[start:i])
				start = i + 1
			}
		}
	}
	if escaped {
		return nil, fmt.Errorf("dangling escape in include-file %s", value)
	}
	if quote != 0 {
		return nil, fmt.Errorf("unclosed quote in include-file %s", value)
	}
	parts = append(parts, value[start:])
	return parts, nil
}

func unquote(value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("delimiter is empty")
	}
	first, _ := utf8Decode(value)
	if first == '"' || first == '\'' {
		if strings.HasSuffix(value, string(first)) && len([]rune(value)) > 1 && !closingQuoteIsEscaped(value) {
			return unescape(value[len(string(first)) : len(value)-len(string(first))])
		}
		return "", fmt.Errorf("delimiter quote is not closed in %s", value)
	}
	return unescape(value)
}

func utf8Decode(value string) (rune, int) {
	for _, r := range value {
		return r, len(string(r))
	}
	return 0, 0
}

func closingQuoteIsEscaped(value string) bool {
	slashes := 0
	runes := []rune(value)
	for i := len(runes) - 2; i >= 0; i-- {
		if runes[i] != '\\' {
			break
		}
		slashes++
	}
	return slashes%2 == 1
}

func unescape(value string) (string, error) {
	var decoded strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] != '\\' {
			decoded.WriteByte(value[i])
			continue
		}
		if i+1 >= len(value) {
			return "", fmt.Errorf("dangling escape in delimiter")
		}
		i++
		decoded.WriteByte(value[i])
	}
	if decoded.Len() == 0 {
		return "", fmt.Errorf("delimiter is empty")
	}
	return decoded.String(), nil
}

func parseColumns(value string) ([]string, error) {
	inner, ok := strings.CutPrefix(value, "[")
	if !ok {
		return nil, fmt.Errorf("columns must be a list in %s", value)
	}
	inner, ok = strings.CutSuffix(inner, "]")
	if !ok {
		return nil, fmt.Errorf("columns must be a list in %s", value)
	}
	var columns []string
	seen := map[string]struct{}{}
	for _, name := range strings.Split(inner, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("column names must be unique")
		}
		seen[name] = struct{}{}
		columns = append(columns, name)
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("columns list is empty")
	}
	return columns, nil
}

func validAlias(alias string) bool {
	if alias == "" {
		return false
	}
	for i := 0; i < len(alias); i++ {
		ch := alias[i]
		letter := (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
		digit := ch >= '0' && ch <= '9'
		if i == 0 {
			if !letter && ch != '_' {
				return false
			}
			continue
		}
		if !letter && !digit && ch != '_' {
			return false
		}
	}
	return true
}
