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

package templater

import (
	"fmt"
	"html"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type node interface {
	render(ev *evaluator) (string, error)
}

type textNode string

func (t textNode) render(*evaluator) (string, error) {
	return string(t), nil
}

type varNode struct {
	path string
}

func (v varNode) render(ev *evaluator) (string, error) {
	value, ok, err := ev.lookup(v.path)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("%s is not defined", v.path)
	}
	text := stringify(value)
	if ev.escapeHTML {
		text = html.EscapeString(text)
	}
	return text, nil
}

type blockNode struct {
	kind     string
	path     string
	body     []node
	elseBody []node
}

func (b blockNode) render(ev *evaluator) (string, error) {
	switch b.kind {
	case "each":
		return b.renderEach(ev)
	case "if", "unless":
		value, ok, err := ev.lookup(b.path)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", fmt.Errorf("%s is not defined", b.path)
		}
		show := truthy(value)
		if b.kind == "unless" {
			show = !show
		}
		if show {
			return renderNodes(b.body, ev)
		}
		return renderNodes(b.elseBody, ev)
	default:
		return "", fmt.Errorf("unknown block %s", b.kind)
	}
}

func (b blockNode) renderEach(ev *evaluator) (string, error) {
	value, ok, err := ev.lookup(b.path)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("%s is not defined", b.path)
	}
	if value == nil {
		return "", nil
	}
	items, err := eachItems(value)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	for _, item := range items {
		child := &scope{
			value: item.value,
			vars: map[string]interface{}{
				"@key":   item.key,
				"@index": item.index,
				"@first": item.index == 0,
				"@last":  item.last,
			},
			parent: ev.scope,
		}
		ev.scope = child
		text, err := renderNodes(b.body, ev)
		ev.scope = child.parent
		if err != nil {
			return "", err
		}
		out.WriteString(text)
	}
	return out.String(), nil
}

type eachItem struct {
	value interface{}
	key   interface{}
	index int
	last  bool
}

func eachItems(value interface{}) ([]eachItem, error) {
	if items, ok := asSlice(value); ok {
		out := make([]eachItem, len(items))
		for i, item := range items {
			out[i] = eachItem{value: item, key: i, index: i, last: i == len(items)-1}
		}
		return out, nil
	}
	if mapping, ok := asMap(value); ok {
		keys := make([]string, 0, len(mapping))
		for key := range mapping {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		out := make([]eachItem, len(keys))
		for i, key := range keys {
			out[i] = eachItem{value: mapping[key], key: key, index: i, last: i == len(keys)-1}
		}
		return out, nil
	}
	return nil, fmt.Errorf("each target is not a collection")
}

type scope struct {
	value  interface{}
	vars   map[string]interface{}
	parent *scope
}

type evaluator struct {
	scope      *scope
	escapeHTML bool
}

func renderBody(body, format string, data interface{}) (string, error) {
	nodes, err := parseTemplate(body)
	if err != nil {
		return "", err
	}
	ev := &evaluator{
		scope:      &scope{value: data},
		escapeHTML: strings.EqualFold(format, "html"),
	}
	return renderNodes(nodes, ev)
}

func renderNodes(nodes []node, ev *evaluator) (string, error) {
	var out strings.Builder
	for _, n := range nodes {
		text, err := n.render(ev)
		if err != nil {
			return "", err
		}
		out.WriteString(text)
	}
	return out.String(), nil
}

func (ev *evaluator) lookup(path string) (interface{}, bool, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, false, fmt.Errorf("missing path")
	}
	if path == "this" || path == "." {
		return ev.scope.value, true, nil
	}
	segments, err := splitPath(path)
	if err != nil {
		return nil, false, err
	}
	if len(segments) == 0 {
		return nil, false, fmt.Errorf("missing path")
	}
	first := segments[0]
	rest := segments[1:]
	if strings.HasPrefix(first, "@") {
		for cur := ev.scope; cur != nil; cur = cur.parent {
			if cur.vars == nil {
				continue
			}
			value, ok := cur.vars[first]
			if !ok {
				continue
			}
			return walk(value, rest)
		}
		return nil, false, nil
	}
	if first == "this" {
		return walk(ev.scope.value, rest)
	}
	for cur := ev.scope; cur != nil; cur = cur.parent {
		base, ok, err := step(cur.value, first)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			continue
		}
		return walk(base, rest)
	}
	return nil, false, nil
}

func walk(value interface{}, segments []string) (interface{}, bool, error) {
	for _, segment := range segments {
		next, ok, err := step(value, segment)
		if err != nil || !ok {
			return nil, ok, err
		}
		value = next
	}
	return value, true, nil
}

func step(value interface{}, segment string) (interface{}, bool, error) {
	if value == nil {
		return nil, false, nil
	}
	if index, ok, err := bracketIndex(segment); err != nil || ok {
		if err != nil {
			return nil, false, err
		}
		return indexSlice(value, index)
	}
	if mapping, ok := asMap(value); ok {
		item, exists := mapping[segment]
		return item, exists, nil
	}
	if index, err := strconv.Atoi(segment); err == nil && index >= 0 {
		if _, isSlice := asSlice(value); isSlice {
			return indexSlice(value, index)
		}
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
		item := rv.MapIndex(reflect.ValueOf(segment))
		if !item.IsValid() {
			return nil, false, nil
		}
		return item.Interface(), true, nil
	}
	return nil, false, nil
}

func bracketIndex(segment string) (int, bool, error) {
	if len(segment) < 3 || segment[0] != '[' || segment[len(segment)-1] != ']' {
		return 0, false, nil
	}
	index, err := strconv.Atoi(segment[1 : len(segment)-1])
	if err != nil || index < 0 {
		return 0, false, fmt.Errorf("invalid index %s", segment)
	}
	return index, true, nil
}

func indexSlice(value interface{}, index int) (interface{}, bool, error) {
	items, ok := asSlice(value)
	if !ok || index >= len(items) {
		return nil, false, nil
	}
	return items[index], true, nil
}

func asSlice(value interface{}) ([]interface{}, bool) {
	switch items := value.(type) {
	case []interface{}:
		return items, true
	case nil:
		return nil, false
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}
	out := make([]interface{}, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		out[i] = rv.Index(i).Interface()
	}
	return out, true
}

func asMap(value interface{}) (map[string]interface{}, bool) {
	mapping, ok := value.(map[string]interface{})
	return mapping, ok
}

func splitPath(path string) ([]string, error) {
	var segments []string
	var current strings.Builder
	flush := func() {
		if current.Len() > 0 {
			segments = append(segments, current.String())
			current.Reset()
		}
	}
	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '.':
			flush()
		case '[':
			flush()
			end := strings.IndexByte(path[i:], ']')
			if end < 0 {
				return nil, fmt.Errorf("invalid path %s", path)
			}
			segments = append(segments, path[i:i+end+1])
			i += end
		default:
			current.WriteByte(path[i])
		}
	}
	flush()
	return segments, nil
}

func parseTemplate(body string) ([]node, error) {
	var nodes []node
	rest := body
	for {
		start := strings.Index(rest, "{{")
		if start < 0 {
			if rest != "" {
				nodes = append(nodes, textNode(rest))
			}
			break
		}
		if start > 0 {
			nodes = append(nodes, textNode(rest[:start]))
		}
		rest = rest[start+2:]
		end := strings.Index(rest, "}}")
		if end < 0 {
			return nil, fmt.Errorf("unclosed template tag")
		}
		tag := strings.TrimSpace(rest[:end])
		rest = rest[end+2:]
		nodes = append(nodes, tagNode(tag))
	}
	return parseBlocks(nodes)
}

func tagNode(tag string) node {
	return rawTag(tag)
}

type rawTag string

func (rawTag) render(*evaluator) (string, error) {
	return "", fmt.Errorf("unparsed template tag")
}

func parseBlocks(nodes []node) ([]node, error) {
	type frame struct {
		kind     string
		path     string
		body     []node
		elseBody []node
		inElse   bool
	}
	root := &frame{}
	stack := []*frame{root}
	for _, n := range nodes {
		tag, ok := n.(rawTag)
		if !ok {
			current := stack[len(stack)-1]
			if current.inElse {
				current.elseBody = append(current.elseBody, n)
			} else {
				current.body = append(current.body, n)
			}
			continue
		}
		text := string(tag)
		switch {
		case text == "else":
			current := stack[len(stack)-1]
			if current.kind != "if" && current.kind != "unless" {
				return nil, fmt.Errorf("unexpected else")
			}
			if current.inElse {
				return nil, fmt.Errorf("unexpected else")
			}
			current.inElse = true
		case strings.HasPrefix(text, "#each "), strings.HasPrefix(text, "#if "), strings.HasPrefix(text, "#unless "):
			kind, path, _ := strings.Cut(text, " ")
			path = strings.TrimSpace(path)
			if path == "" {
				return nil, fmt.Errorf("missing path")
			}
			stack = append(stack, &frame{kind: strings.TrimPrefix(kind, "#"), path: path})
		case text == "/each" || text == "/if" || text == "/unless":
			if len(stack) == 1 {
				return nil, fmt.Errorf("unexpected %s", text)
			}
			current := stack[len(stack)-1]
			closer := strings.TrimPrefix(text, "/")
			if current.kind != closer {
				return nil, fmt.Errorf("unexpected %s", text)
			}
			stack = stack[:len(stack)-1]
			parent := stack[len(stack)-1]
			block := blockNode{kind: current.kind, path: current.path, body: current.body, elseBody: current.elseBody}
			if parent.inElse {
				parent.elseBody = append(parent.elseBody, block)
			} else {
				parent.body = append(parent.body, block)
			}
		default:
			if strings.HasPrefix(text, "#") || strings.HasPrefix(text, "/") {
				return nil, fmt.Errorf("unknown template tag %s", text)
			}
			current := stack[len(stack)-1]
			variable := varNode{path: text}
			if current.inElse {
				current.elseBody = append(current.elseBody, variable)
			} else {
				current.body = append(current.body, variable)
			}
		}
	}
	if len(stack) != 1 {
		return nil, fmt.Errorf("unclosed %s", stack[len(stack)-1].kind)
	}
	return root.body, nil
}

func truthy(value interface{}) bool {
	if value == nil {
		return false
	}
	switch n := value.(type) {
	case bool:
		return n
	case string:
		return n != ""
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		return rv.Len() != 0
	case reflect.Bool:
		return rv.Bool()
	default:
		return true
	}
}

func stringify(value interface{}) string {
	switch n := value.(type) {
	case nil:
		return ""
	case string:
		return n
	case bool:
		return strconv.FormatBool(n)
	case float64:
		if !math.IsInf(n, 0) && n == math.Trunc(n) {
			return strconv.FormatInt(int64(n), 10)
		}
		return strconv.FormatFloat(n, 'f', -1, 64)
	case float32:
		return stringify(float64(n))
	default:
		return fmt.Sprint(n)
	}
}
