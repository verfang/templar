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
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/verfang/templater/mapper"
)

func writeTemp(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func requireValue(t *testing.T, m mapper.Mapper, path, key string, want interface{}) {
	t.Helper()
	if err := m.LoadFile(path); err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	requireLoaded(t, m, key, want)
}

func requireLoaded(t *testing.T, m mapper.Mapper, key string, want interface{}) {
	t.Helper()
	got, err := m.GetProperty(key)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetProperty(%q) = %#v, want %#v", key, got, want)
	}
}

func requireMissing(t *testing.T, m mapper.Mapper, key string) {
	t.Helper()
	if _, err := m.GetProperty(key); err == nil {
		t.Fatalf("expected property %q to be missing", key)
	}
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	data, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
