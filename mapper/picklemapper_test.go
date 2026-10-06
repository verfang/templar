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

func TestPickleMapperLoadsDict(t *testing.T) {
	data := mustHex(t, "8004952d000000000000007d94288c046e616d65948c03416461948c03616765944b248c0474616773945d94288c0161948c01629465752e")
	path := writeTemp(t, "cfg.pkl", data)
	m := mapper.NewPickleMapper()
	requireValue(t, m, path, "name", "Ada")
	requireLoaded(t, m, "age", int64(36))
	requireLoaded(t, m, "tags", []interface{}{"a", "b"})
}

func TestPickleMapperLoadsNestedDict(t *testing.T) {
	data := mustHex(t, "80049520000000000000007d948c06706572736f6e947d948c0463697479948c064c6f6e646f6e9473732e")
	path := writeTemp(t, "cfg.pkl", data)
	requireValue(t, mapper.NewPickleMapper(), path, "person", map[string]interface{}{"city": "London"})
}

func TestPickleMapperSetAndMissingProperty(t *testing.T) {
	data := mustHex(t, "8004952d000000000000007d94288c046e616d65948c03416461948c03616765944b248c0474616773945d94288c0161948c01629465752e")
	path := writeTemp(t, "cfg.pkl", data)
	m := mapper.NewPickleMapper()
	if err := m.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	if err := m.SetProperty("name", "Grace"); err != nil {
		t.Fatal(err)
	}
	requireLoaded(t, m, "name", "Grace")
	requireMissing(t, m, "missing")
}

func TestPickleMapperRejectsNonDict(t *testing.T) {
	data := mustHex(t, "8004950b000000000000005d948c046e6f706594612e")
	path := writeTemp(t, "cfg.pkl", data)
	if err := mapper.NewPickleMapper().LoadFile(path); err == nil {
		t.Fatal("expected non-dict pickle to fail")
	}
}
