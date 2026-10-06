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
	"bytes"
	"fmt"
	"math/big"

	ogorek "github.com/kisielk/og-rek"
)

type PickleMapper struct {
	store
}

func NewPickleMapper() *PickleMapper {
	return &PickleMapper{store: newStore()}
}

func (m *PickleMapper) LoadFile(path string) error {
	return m.loadDocument(path, decodePickle)
}

func decodePickle(data []byte) (map[string]interface{}, error) {
	value, err := ogorek.NewDecoder(bytes.NewReader(data)).Decode()
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	normalized, err := normalizePickle(value)
	if err != nil {
		return nil, err
	}
	doc, ok := normalized.(map[string]interface{})
	if !ok || normalized == nil {
		return nil, fmt.Errorf("failed to read file: pickle value must be a dict with string keys")
	}
	return doc, nil
}

func normalizePickle(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case nil, string, bool, int, int64, float64, []byte:
		return v, nil
	case *big.Int:
		return v, nil
	case ogorek.None:
		return nil, nil
	case ogorek.Bytes:
		return []byte(v), nil
	case ogorek.ByteString:
		return string(v), nil
	case ogorek.Tuple:
		return normalizePickle([]any(v))
	case []any:
		out := make([]interface{}, len(v))
		for i, item := range v {
			normalized, err := normalizePickle(item)
			if err != nil {
				return nil, err
			}
			out[i] = normalized
		}
		return out, nil
	case map[any]any:
		out := make(map[string]interface{}, len(v))
		for key, item := range v {
			name, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("failed to read file: pickle key %v is not a string", key)
			}
			normalized, err := normalizePickle(item)
			if err != nil {
				return nil, err
			}
			out[name] = normalized
		}
		return out, nil
	case ogorek.Call:
		return nil, fmt.Errorf("failed to read file: pickle value contains Python object %s.%s", v.Callable.Module, v.Callable.Name)
	case ogorek.Class:
		return nil, fmt.Errorf("failed to read file: pickle value contains Python class %s.%s", v.Module, v.Name)
	case ogorek.Ref:
		return nil, fmt.Errorf("failed to read file: pickle persistent id is not supported")
	default:
		return nil, fmt.Errorf("failed to read file: pickle value %T is not supported", value)
	}
}

var _ Mapper = (*PickleMapper)(nil)
