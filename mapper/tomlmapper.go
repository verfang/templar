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

import "github.com/pelletier/go-toml/v2"

type TOMLMapper struct {
	store
}

func NewTOMLMapper() *TOMLMapper {
	return &TOMLMapper{store: newStore()}
}

func (m *TOMLMapper) LoadFile(path string) error {
	return m.loadDocument(path, func(data []byte) (map[string]interface{}, error) {
		return unmarshalObject(data, toml.Unmarshal, true)
	})
}

var _ Mapper = (*TOMLMapper)(nil)
