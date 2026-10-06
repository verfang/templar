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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/verfang/templer/mapper"
)

func Render(templatePath string) (string, error) {
	template, err := os.ReadFile(templatePath)
	if err != nil {
		return "", err
	}
	prepared, err := prepareTemplate(string(template))
	if err != nil {
		return "", err
	}
	outputDir, ok := prepared.metadata["output-dir"]
	if !ok {
		return "", fmt.Errorf("template is missing !output-dir")
	}
	outputFile, ok := prepared.metadata["output-file"]
	if !ok {
		return "", fmt.Errorf("template is missing !output-file")
	}
	var validateFormat string
	if format, exists := prepared.metadata["validate_format"]; exists {
		if format == "" {
			return "", fmt.Errorf("validate_format is empty")
		}
		validateFormat = format
	}
	if len(prepared.includeFiles) == 0 {
		return "", fmt.Errorf("template is missing !include-file")
	}

	data := map[string]interface{}{}
	aliases := map[string]struct{}{}
	for _, include := range prepared.includeFiles {
		spec, err := parseInclude(include)
		if err != nil {
			return "", err
		}
		if _, exists := aliases[spec.alias]; exists {
			return "", fmt.Errorf("duplicate include alias %s", spec.alias)
		}
		aliases[spec.alias] = struct{}{}
		path := resolveInclude(templatePath, spec.path)
		loaded, err := loadInclude(path, spec)
		if err != nil {
			return "", err
		}
		data[spec.alias] = loaded
	}

	rendered, err := renderBody(prepared.body, validateFormat, data)
	if err != nil {
		return "", err
	}
	if validateFormat != "" {
		if err := checkFormat(validateFormat, rendered); err != nil {
			return "", err
		}
	}

	destination := resolveOutput(templatePath, outputDir, outputFile)
	if parent := filepath.Dir(destination); parent != "" {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return "", err
		}
	}
	if err := os.WriteFile(destination, []byte(rendered), 0o644); err != nil {
		return "", err
	}
	return destination, nil
}

func loadInclude(path string, include includeFile) (interface{}, error) {
	if strings.EqualFold(filepath.Ext(path), ".csv") {
		if len(include.columns) == 0 {
			return nil, fmt.Errorf("csv include %s is missing columns", include.path)
		}
		delimiter := ","
		if include.delimiter != nil {
			delimiter = *include.delimiter
		}
		text, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		body, err := prepareSource(string(text))
		if err != nil {
			return nil, err
		}
		return delimitedBody(body, delimiter, include.columns)
	}
	if len(include.columns) > 0 || include.delimiter != nil {
		return nil, fmt.Errorf("columns and delimiter belong on a csv include, not %s", include.path)
	}
	return loadDocument(path)
}

func loadDocument(path string) (interface{}, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return decodeFile(path, func(data []byte) (interface{}, error) {
			var value interface{}
			err := json.Unmarshal(data, &value)
			return value, err
		})
	case ".toml":
		return decodeFile(path, func(data []byte) (interface{}, error) {
			var value interface{}
			err := toml.Unmarshal(data, &value)
			return value, err
		})
	case ".yaml", ".yml":
		return decodeFile(path, func(data []byte) (interface{}, error) {
			var value interface{}
			err := yaml.Unmarshal(data, &value)
			return value, err
		})
	case ".xml":
		return loadMapped(path, mapper.NewXMLMapper())
	case ".properties":
		return loadMapped(path, mapper.NewPropertiesMapper())
	case ".pickle", ".pkl":
		return loadMapped(path, mapper.NewPickleMapper())
	default:
		return nil, fmt.Errorf("unsupported config extension %s", path)
	}
}

func decodeFile(path string, decode func([]byte) (interface{}, error)) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	value, err := decode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return value, nil
}

func loadMapped(path string, doc mapper.Mapper) (interface{}, error) {
	if err := doc.LoadFile(path); err != nil {
		return nil, err
	}
	return doc.Map(), nil
}

func delimitedBody(body, delimiter string, columns []string) ([]interface{}, error) {
	var rows []interface{}
	rowNumber := 0
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		rowNumber++
		fields := splitRow(line, delimiter)
		if len(fields) != len(columns) {
			return nil, fmt.Errorf("row %d has %d fields, expected %d", rowNumber, len(fields), len(columns))
		}
		record := make(map[string]interface{}, len(columns))
		for i, name := range columns {
			if fields[i] == "" {
				record[name] = nil
				continue
			}
			record[name] = fields[i]
		}
		rows = append(rows, record)
	}
	return rows, nil
}

func splitRow(line, delimiter string) []string {
	fields := strings.Split(line, delimiter)
	if strings.HasSuffix(line, delimiter) && len(fields) > 0 && fields[len(fields)-1] == "" {
		fields = fields[:len(fields)-1]
	}
	return fields
}

func resolveInclude(templatePath, include string) string {
	if filepath.IsAbs(include) {
		return include
	}
	return filepath.Join(filepath.Dir(templatePath), include)
}

func resolveOutput(templatePath, outputDir, outputFile string) string {
	if filepath.IsAbs(outputFile) {
		return outputFile
	}
	directory := outputDir
	if !filepath.IsAbs(directory) {
		directory = filepath.Join(filepath.Dir(templatePath), directory)
	}
	return filepath.Join(directory, outputFile)
}

func checkFormat(format, rendered string) error {
	var err error
	switch strings.ToLower(format) {
	case "json":
		var value interface{}
		err = json.Unmarshal([]byte(rendered), &value)
	case "yaml", "yml":
		var value interface{}
		err = yaml.Unmarshal([]byte(rendered), &value)
	case "toml":
		var value interface{}
		err = toml.Unmarshal([]byte(rendered), &value)
	case "html", "xml":
		return nil
	default:
		err = fmt.Errorf("unsupported output format %s", format)
	}
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
