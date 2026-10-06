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

func TestXMLMapperLoadsChildren(t *testing.T) {
	path := writeTemp(t, "cfg.xml", []byte("<config><name>Ada</name><person><city>London</city></person></config>\n"))
	m := mapper.NewXMLMapper()
	requireValue(t, m, path, "name", "Ada")
	requireLoaded(t, m, "person", map[string]interface{}{"city": "London"})
}

func TestXMLMapperAttributesAndRepeatedElements(t *testing.T) {
	body := "<config id=\"1\"><name>Ada</name><tag>a</tag><tag>b</tag></config>\n"
	path := writeTemp(t, "cfg.xml", []byte(body))
	m := mapper.NewXMLMapper()
	if err := m.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	requireLoaded(t, m, "@id", "1")
	requireLoaded(t, m, "name", "Ada")
	requireLoaded(t, m, "tag", []interface{}{"a", "b"})
}

func TestXMLMapperTextRoot(t *testing.T) {
	path := writeTemp(t, "cfg.xml", []byte("<name>Ada</name>\n"))
	requireValue(t, mapper.NewXMLMapper(), path, "name", "Ada")
}

func TestXMLMapperSetAndMissingProperty(t *testing.T) {
	path := writeTemp(t, "cfg.xml", []byte("<config><name>Ada</name></config>\n"))
	m := mapper.NewXMLMapper()
	if err := m.LoadFile(path); err != nil {
		t.Fatal(err)
	}
	if err := m.SetProperty("name", "Grace"); err != nil {
		t.Fatal(err)
	}
	requireLoaded(t, m, "name", "Grace")
	requireMissing(t, m, "missing")
}

func TestXMLMapperRejectsText(t *testing.T) {
	path := writeTemp(t, "cfg.xml", []byte("not xml\n"))
	if err := mapper.NewXMLMapper().LoadFile(path); err == nil {
		t.Fatal("expected invalid xml to fail")
	}
}
