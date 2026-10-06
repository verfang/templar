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

	"github.com/verfang/templater/mapper"
)

func TestTOMLMapperLoadsDocument(t *testing.T) {
	body := "name = \"Ada\"\nage = 36\ntags = [\"a\", \"b\"]\n\n[person]\ncity = \"London\"\n"
	path := writeTemp(t, "cfg.toml", []byte(body))
	m := mapper.NewTOMLMapper()
	requireValue(t, m, path, "name", "Ada")
	requireLoaded(t, m, "person", map[string]interface{}{"city": "London"})
}

func TestTOMLMapperSetAndMissingProperty(t *testing.T) {
	path := writeTemp(t, "cfg.toml", []byte("name = \"Ada\"\n"))
	m := mapper.NewTOMLMapper()
	if err := m.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	if err := m.SetProperty("name", "Grace"); err != nil {
		t.Fatal(err)
	}
	requireLoaded(t, m, "name", "Grace")
	requireMissing(t, m, "missing")
}

func TestTOMLMapperRejectsInvalidDocument(t *testing.T) {
	path := writeTemp(t, "cfg.toml", []byte("=\n"))
	if err := mapper.NewTOMLMapper().LoadFile(path); err == nil {
		t.Fatal("expected invalid toml to fail")
	}
}

func TestTOMLMapperEmptyFile(t *testing.T) {
	path := writeTemp(t, "cfg.toml", nil)
	m := mapper.NewTOMLMapper()
	if err := m.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	requireMissing(t, m, "name")
}
