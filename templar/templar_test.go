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

package templar

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWritesRenderedJSONIntoTheOutputDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "child.json"), `{"example":"example value","map":{"key":"value","key2":"value2"}}`)
	writeFile(t, filepath.Join(dir, "data.toml"), "name = \"testing\"\nmulti_args = [1, 2]\nmuli_spaced = \"1 2 3\"\n")
	writeFile(t, filepath.Join(dir, "people.csv"), "richie|wood|tn|\n")
	writeFile(t, filepath.Join(dir, "view.template"), `# !output-dir: output
# !output-file: final.json
# !validate_format: json
# !include-file: data.toml, as: t1
# !include-file: child.json, as: t2
# !include-file: people.csv, delimiter: |, columns: [first_name, last_name, state], as: t3
{
  "name": "{{t1.name}}",
  "admin": "{{#if t1.name}}yes{{else}}no{{/if}}",
  "multi_args": [{{#each t1.multi_args}}{{this}}{{#unless @last}},{{/unless}}{{/each}}],
  "muli_spaced": "{{t1.muli_spaced}}",
  "example": "{{t2.example}}",
  "first": "{{t3.[0].first_name}}",
  "state": "{{t3.0.state}}",
  "map": {
    {{#each t2.map}}"{{@key}}": "{{this}}"{{#unless @last}},{{/unless}}{{/each}}
  }
}
`)

	written, err := Render(filepath.Join(dir, "view.template"))
	if err != nil {
		t.Fatal(err)
	}
	if written != filepath.Join(dir, "output", "final.json") {
		t.Fatalf("wrote %s", written)
	}

	var value map[string]interface{}
	text, err := os.ReadFile(written)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(text, &value); err != nil {
		t.Fatalf("rendered json: %v\n%s", err, text)
	}
	if value["name"] != "testing" || value["admin"] != "yes" || value["muli_spaced"] != "1 2 3" || value["example"] != "example value" {
		t.Fatalf("document = %#v", value)
	}
	if value["first"] != "richie" || value["state"] != "tn" {
		t.Fatalf("csv fields = %#v", value)
	}
	args, ok := value["multi_args"].([]interface{})
	if !ok || len(args) != 2 || args[0].(float64) != 1 || args[1].(float64) != 2 {
		t.Fatalf("multi_args = %#v", value["multi_args"])
	}
	mapped, ok := value["map"].(map[string]interface{})
	if !ok || mapped["key2"] != "value2" {
		t.Fatalf("map = %#v", value["map"])
	}
}

func TestAcceptsQuotedOrEscapedCommaDelimiter(t *testing.T) {
	quoted, err := parseInclude(`people.csv, delimiter: ",", columns: [first_name, last_name], as: people`)
	if err != nil {
		t.Fatal(err)
	}
	if quoted.delimiter == nil || *quoted.delimiter != "," {
		t.Fatalf("quoted delimiter = %#v", quoted.delimiter)
	}

	escaped, err := parseInclude(`people.csv, delimiter: \,, columns: [first_name, last_name], as: people`)
	if err != nil {
		t.Fatal(err)
	}
	if escaped.delimiter == nil || *escaped.delimiter != "," {
		t.Fatalf("escaped delimiter = %#v", escaped.delimiter)
	}

	single, err := parseInclude("people.csv, delimiter: ',', columns: [first_name, last_name], as: people")
	if err != nil {
		t.Fatal(err)
	}
	if single.delimiter == nil || *single.delimiter != "," {
		t.Fatalf("single quoted delimiter = %#v", single.delimiter)
	}

	pipes, err := parseInclude(`people.csv, delimiter: "|", columns: [first_name], as: people`)
	if err != nil {
		t.Fatal(err)
	}
	if pipes.delimiter == nil || *pipes.delimiter != "|" {
		t.Fatalf("pipe delimiter = %#v", pipes.delimiter)
	}

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "people.csv"), "richie,wood\n")
	writeFile(t, filepath.Join(dir, "view.template"), "# !output-dir: output\n# !output-file: final.txt\n# !include-file: people.csv, delimiter: \",\", columns: [first_name, last_name], as: t3\n{{t3.[0].first_name}}\n")
	written, err := Render(filepath.Join(dir, "view.template"))
	if err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(written)
	if err != nil {
		t.Fatal(err)
	}
	if string(text) != "richie\n" {
		t.Fatalf("rendered %q", text)
	}
}

func TestValidationErrorDoesNotWriteTheFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "data.json"), `{"name":"testing"}`)
	writeFile(t, filepath.Join(dir, "view.template"), "# !output-dir: output\n# !output-file: final.json\n# !validate_format: json\n# !include-file: data.json, as: t1\nnot json {{t1.name}}\n")
	if _, err := Render(filepath.Join(dir, "view.template")); err == nil {
		t.Fatal("expected validation to fail")
	}
	if _, err := os.Stat(filepath.Join(dir, "output", "final.json")); !os.IsNotExist(err) {
		t.Fatal("invalid output was written")
	}
}

func TestWritesWithoutCheckingWhenValidateFormatIsAbsent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "data.json"), `{"name":"testing"}`)
	writeFile(t, filepath.Join(dir, "view.template"), "# !output-dir: output\n# !output-file: final.txt\n# !include-file: data.json, as: t1\nnot json {{t1.name}}\n")
	written, err := Render(filepath.Join(dir, "view.template"))
	if err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(written)
	if err != nil {
		t.Fatal(err)
	}
	if string(text) != "not json testing\n" {
		t.Fatalf("rendered %q", text)
	}
}

func TestRequiresAnOutputDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "view.template"), "# !output-file: final.json\n{\"name\":\"{{name}}\"}\n")
	_, err := Render(filepath.Join(dir, "view.template"))
	if err == nil || !strings.Contains(err.Error(), "output-dir") {
		t.Fatalf("error = %v", err)
	}
}

func TestKeepsHandlebarsMarkersAndEachInclude(t *testing.T) {
	prepared, err := prepareTemplate("# !output-format: json\n\"items\": [{{#each items}}{{this}}{{/each}}]\n")
	if err != nil {
		t.Fatal(err)
	}
	if prepared.metadata["output-format"] != "json" {
		t.Fatalf("metadata = %#v", prepared.metadata)
	}
	if prepared.body != "\"items\": [{{#each items}}{{this}}{{/each}}]\n" {
		t.Fatalf("body = %q", prepared.body)
	}

	includes, err := prepareTemplate("# !include-file: test.toml, as: t1\n# !include-file: test2.json, as: t2\n{}\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(includes.includeFiles) != 2 || includes.includeFiles[0] != "test.toml, as: t1" || includes.includeFiles[1] != "test2.json, as: t2" {
		t.Fatalf("includes = %#v", includes.includeFiles)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
