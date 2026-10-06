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

	"github.com/verfang/templer/mapper"
)

func TestPropertiesMapperEscapes(t *testing.T) {
	body := "# comment\n" +
		"! another\n" +
		"name = Ada\n" +
		"city: London\n" +
		"note this is a note\n" +
		"path=C:\\\\temp\n" +
		"snow=\\u2603\n" +
		"long=hello \\\n" +
		"\tworld\n" +
		"a\\=b=c\n"
	path := writeTemp(t, "cfg.properties", []byte(body))
	m := mapper.NewPropertiesMapper()
	if err := m.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	requireLoaded(t, m, "name", "Ada")
	requireLoaded(t, m, "city", "London")
	requireLoaded(t, m, "note", "this is a note")
	requireLoaded(t, m, "path", `C:\temp`)
	requireLoaded(t, m, "snow", "☃")
	requireLoaded(t, m, "long", "hello world")
	requireLoaded(t, m, "a=b", "c")
}

func TestPropertiesMapperSetAndMissingProperty(t *testing.T) {
	path := writeTemp(t, "cfg.properties", []byte("name=Ada\n"))
	m := mapper.NewPropertiesMapper()
	if err := m.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	if err := m.SetProperty("name", "Grace"); err != nil {
		t.Fatal(err)
	}
	requireLoaded(t, m, "name", "Grace")
	requireMissing(t, m, "missing")
}

func TestPropertiesMapperRejectsBadEscape(t *testing.T) {
	path := writeTemp(t, "cfg.properties", []byte("snow=\\u12\n"))
	if err := mapper.NewPropertiesMapper().LoadFile(path); err == nil {
		t.Fatal("expected invalid unicode escape to fail")
	}
}
