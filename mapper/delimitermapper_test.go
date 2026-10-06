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
	"os"
	"path/filepath"
	"testing"

	"github.com/verfang/templar/mapper"
)

func TestDelimiterMapperLoadsNamedColumns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.csv")
	if err := os.WriteFile(path, []byte("Ada,36\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := mapper.NewDelimiterMapper(",", []string{"name", "age"})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.LoadFile(path); err != nil {
		t.Fatal(err)
	}

	got, err := m.GetProperty("name")
	if err != nil {
		t.Fatal(err)
	}
	if got != "Ada" {
		t.Fatalf("GetProperty(name) = %v, want Ada", got)
	}
}

func TestDelimiterMapperUsesTheDelimiterAndKeepsTheLastRow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.psv")
	if err := os.WriteFile(path, []byte("Ada|36\nGrace|40\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := mapper.NewDelimiterMapper("|", []string{"name", "age"})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	requireLoaded(t, m, "name", "Grace")
	requireLoaded(t, m, "age", "40")
	if err := m.SetProperty("name", "Edsger"); err != nil {
		t.Fatal(err)
	}
	requireLoaded(t, m, "name", "Edsger")
	requireMissing(t, m, "city")
}

func TestDelimiterMapperRejectsABadDelimiterOrShortRow(t *testing.T) {
	if _, err := mapper.NewDelimiterMapper("", []string{"name"}); err == nil {
		t.Fatal("expected an empty delimiter to fail")
	}
	if _, err := mapper.NewDelimiterMapper("||", []string{"name"}); err == nil {
		t.Fatal("expected a multi-character delimiter to fail")
	}

	path := filepath.Join(t.TempDir(), "data.csv")
	if err := os.WriteFile(path, []byte("Ada\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := mapper.NewDelimiterMapper(",", []string{"name", "age"})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.LoadFile(path); err == nil {
		t.Fatal("expected a short row to fail")
	}
}
