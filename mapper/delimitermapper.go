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
	"encoding/csv"
	"fmt"
	"os"
)

type DelimiterMapper struct {
	delimiter rune
	columns   []string
	store
}

func NewDelimiterMapper(delimiter string, columns []string) (*DelimiterMapper, error) {
	runes := []rune(delimiter)
	if len(runes) != 1 {
		return nil, fmt.Errorf("delimiter must be a single character")
	}
	return &DelimiterMapper{
		delimiter: runes[0],
		columns:   columns,
		store:     newStore(),
	}, nil
}

func (m *DelimiterMapper) LoadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = m.delimiter
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	for _, row := range records {
		if len(row) < len(m.columns) {
			return fmt.Errorf("record has %d fields, want at least %d", len(row), len(m.columns))
		}
		for i, col := range m.columns {
			m.Properties[col] = row[i]
		}
	}
	return nil
}

var _ Mapper = (*DelimiterMapper)(nil)
