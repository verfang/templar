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
	"os"
)

type store struct {
	Properties map[string]interface{}
}

func newStore() store {
	return store{Properties: make(map[string]interface{})}
}

func (s *store) GetProperty(key string) (interface{}, error) {
	value, ok := s.Properties[key]
	if !ok {
		return nil, fmt.Errorf("property %q not found", key)
	}
	return value, nil
}

func (s *store) Map() map[string]interface{} {
	if s.Properties == nil {
		return map[string]interface{}{}
	}
	return s.Properties
}

func (s *store) SetProperty(key string, value interface{}) error {
	if s.Properties == nil {
		s.Properties = make(map[string]interface{})
	}
	s.Properties[key] = value
	return nil
}

func (s *store) loadDocument(path string, decode func([]byte) (map[string]interface{}, error)) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	doc, err := decode(data)
	if err != nil {
		return err
	}
	s.Properties = doc
	return nil
}

func unmarshalObject(data []byte, unmarshal func([]byte, any) error, emptyIfNil bool) (map[string]interface{}, error) {
	var doc map[string]interface{}
	if err := unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if doc == nil {
		if !emptyIfNil {
			return nil, fmt.Errorf("failed to read file: value must be an object")
		}
		doc = map[string]interface{}{}
	}
	return doc, nil
}
