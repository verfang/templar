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

package mapper_test

import (
	"testing"

	"github.com/verfang/templar/mapper"
)

func TestJSONMapperLoadsDocument(t *testing.T) {
	path := writeTemp(t, "cfg.json", []byte("{\"name\":\"Ada\",\"age\":36,\"admin\":true,\"tags\":[\"a\",\"b\"],\"person\":{\"city\":\"London\"}}\n"))
	m := mapper.NewJSONMapper()
	requireValue(t, m, path, "name", "Ada")
	requireLoaded(t, m, "age", float64(36))
	requireLoaded(t, m, "admin", true)
	requireLoaded(t, m, "tags", []interface{}{"a", "b"})
	requireLoaded(t, m, "person", map[string]interface{}{"city": "London"})
	if m.Map()["name"] != "Ada" {
		t.Fatalf("Map() = %#v", m.Map())
	}
}

func TestJSONMapperSetAndMissingProperty(t *testing.T) {
	path := writeTemp(t, "cfg.json", []byte("{\"name\":\"Ada\"}\n"))
	m := mapper.NewJSONMapper()
	if err := m.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	if err := m.SetProperty("name", "Grace"); err != nil {
		t.Fatal(err)
	}
	requireLoaded(t, m, "name", "Grace")
	requireMissing(t, m, "missing")
}

func TestJSONMapperRejectsNonObjects(t *testing.T) {
	for _, body := range []string{"[1]\n", "null\n", "{\n"} {
		path := writeTemp(t, "cfg.json", []byte(body))
		if err := mapper.NewJSONMapper().LoadFile(path); err == nil {
			t.Fatalf("expected %q to fail", body)
		}
	}
}

func TestJSONMapperMissingFile(t *testing.T) {
	if err := mapper.NewJSONMapper().LoadFile(writeTemp(t, "missing", nil) + "-absent.json"); err == nil {
		t.Fatal("expected missing file to fail")
	}
}
