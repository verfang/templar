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
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type XMLMapper struct {
	store
}

func NewXMLMapper() *XMLMapper {
	return &XMLMapper{store: newStore()}
}

func (m *XMLMapper) LoadFile(path string) error {
	return m.loadDocument(path, decodeXML)
}

type xmlElement struct {
	name     string
	attrs    []xml.Attr
	text     string
	children []*xmlElement
}

func decodeXML(data []byte) (map[string]interface{}, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil, fmt.Errorf("failed to read file: xml document is empty")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		root, err := readXMLElement(dec, start)
		if err != nil {
			return nil, err
		}
		return xmlProperties(root)
	}
}

func readXMLElement(dec *xml.Decoder, start xml.StartElement) (*xmlElement, error) {
	el := &xmlElement{
		name:  start.Name.Local,
		attrs: append([]xml.Attr(nil), start.Attr...),
	}
	var text strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			child, err := readXMLElement(dec, t)
			if err != nil {
				return nil, err
			}
			el.children = append(el.children, child)
		case xml.EndElement:
			el.text = text.String()
			return el, nil
		case xml.CharData:
			text.Write(t)
		case xml.Comment, xml.ProcInst, xml.Directive:
		default:
			return nil, fmt.Errorf("failed to read file: unexpected xml token %T", tok)
		}
	}
}

func xmlProperties(el *xmlElement) (map[string]interface{}, error) {
	if len(el.attrs) == 0 && len(el.children) == 0 {
		return map[string]interface{}{el.name: strings.TrimSpace(el.text)}, nil
	}
	value, ok := xmlValue(el).(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to read file: xml value must be an object")
	}
	return value, nil
}

func xmlValue(el *xmlElement) interface{} {
	text := strings.TrimSpace(el.text)
	if len(el.attrs) == 0 && len(el.children) == 0 {
		return text
	}

	out := make(map[string]interface{})
	for _, attr := range el.attrs {
		out["@"+attr.Name.Local] = attr.Value
	}

	grouped := make(map[string][]*xmlElement)
	var order []string
	for _, child := range el.children {
		if _, seen := grouped[child.name]; !seen {
			order = append(order, child.name)
		}
		grouped[child.name] = append(grouped[child.name], child)
	}
	for _, name := range order {
		group := grouped[name]
		if len(group) == 1 {
			out[name] = xmlValue(group[0])
			continue
		}
		values := make([]interface{}, len(group))
		for i, child := range group {
			values[i] = xmlValue(child)
		}
		out[name] = values
	}
	if text != "" {
		out["#text"] = text
	}
	return out
}

var _ Mapper = (*XMLMapper)(nil)
