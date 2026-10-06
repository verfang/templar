# templater

Copyright 2026 Richie Wood

templater loads env files in several formats and renders a template that ties
those files together. The template names every input, the output path, and an
optional check of the rendered text. Env files stay single documents. They do
not pull in other files.

The template body is a Handlebars subset: `{{name}}`, `{{#if}}`, `{{else}}`,
`{{#unless}}`, and `{{#each}}`, including `{{this}}`, `{{@key}}`, `{{@index}}`,
`{{@first}}`, and `{{@last}}`. A missing name fails the render.

    https://handlebarsjs.com/

The module is Apache-2.0. See LICENSE.


Build
-----

From the repository root:

    go test ./...


What a template does
--------------------

A template is a text file. Lines whose first non-whitespace character is #
are directives or comments. The rest of the file is the template body, so
`{{#each}}` and `{{#if}}` stay in the body.

Required directives:

    # !output-dir: output
    # !output-file: final.json
    # !include-file: app.toml, as: app

`output-dir` and `output-file` may be absolute, or relative to the template's
directory. The output directory is created only after the template renders
and any format check succeeds.

Repeat `include-file` once per env file. Each line needs an alias after `as:`.
The alias is a letter or underscore, then letters, digits, or underscores.
The same alias twice is an error. At least one include is required. Paths are
relative to the template's directory unless they are absolute.

Optional:

    # !validate_format: json

When this is set, the rendered text is checked before anything is written.
`json`, `yaml`, `yml`, and `toml` must parse. `html` and `xml` are accepted
without a parse check. `html` is also the only value that HTML-escapes
rendered values. Any other value, including an empty one, is an error, and
the output file is not written. Leave the directive off to write the text as
rendered.

Supported include extensions:

    .json
    .yaml  .yml
    .toml
    .xml
    .csv
    .properties
    .pkl  .pickle

A csv include also names its columns, and may name a delimiter. The default
delimiter is a comma. `columns` and `delimiter` on any other include are an
error.

    # !include-file: people.csv, columns: [first_name, last_name, state], as: people
    # !include-file: people.csv, delimiter: "|", columns: [first_name, last_name, state], as: people

A comma or a quote in the delimiter has to be quoted or escaped, because a
bare comma splits the include line:

    delimiter: ","
    delimiter: ','
    delimiter: \,
    delimiter: "'"
    delimiter: \'
    delimiter: "\""
    delimiter: \"

A csv row is one record. Empty fields become null. A trailing delimiter after
the last value is ignored, so `tn|` and `tn` are the same row when the last
column has a value. The field count must match the column list. `#` comments
in the csv file are removed before the rows are read.

Inside the template, each alias is the included document. A csv alias is an
array of row objects.

    {{app.name}}
    {{#each app.ports}}{{this}}{{#unless @last}}, {{/unless}}{{/each}}
    {{people.[0].first_name}}
    {{people.0.state}}
    {{#each people}}{{first_name}} {{../app.name}}{{/each}}

`{{people[0].first_name}}` is not valid. Use a dot before the bracket index,
or a bare numeric segment, as above. Inside `{{#each}}`, other aliases are
still in scope. Missing names fail the render.


Go
--

Render a template. `Render` returns the path it wrote.

    package main

    import (
        "fmt"

        "github.com/verfang/templater/templater"
    )

    func main() {
        written, err := templater.Render("view.template")
        if err != nil {
            fmt.Println(err)
            return
        }
        fmt.Println(written)
    }

`view.template`:

    # !output-dir: output
    # !output-file: final.json
    # !validate_format: json
    # !include-file: app.toml, as: app
    # !include-file: people.csv, delimiter: "|", columns: [first_name, last_name, state], as: people
    {
      "name": "{{app.name}}",
      "port": {{app.port}},
      "first": "{{people.[0].first_name}}"
    }

`app.toml`:

    name = "warehouse"
    port = 5432

`people.csv`:

    jordan|hale|tn|

That writes `output/final.json` next to the template:

    {
      "name": "warehouse",
      "port": 5432,
      "first": "jordan"
    }

If `validate_format` fails, `Render` returns the error and leaves the output
path unwritten.

Load one env file when you want the data without rendering. The constructor
selects the format.

    package main

    import (
        "fmt"

        "github.com/verfang/templater/mapper"
    )

    func main() {
        file := mapper.NewJSONMapper()
        if err := file.LoadFile("app.json"); err != nil {
            fmt.Println(err)
            return
        }
        value, err := file.GetProperty("name")
        if err != nil {
            fmt.Println(err)
            return
        }
        fmt.Println(value)
        if err := file.SetProperty("name", "other"); err != nil {
            fmt.Println(err)
            return
        }
    }

`GetProperty` reads one top-level key and returns the decoded value.
A missing key is an error. `SetProperty` changes the loaded value in memory.
It does not write the file.

`NewDelimiterMapper` reads a delimited file with an explicit delimiter and
column list. `GetProperty` returns the named column from the last row.
A template `include-file` for `.csv` does not use that mapper. It uses the
`columns` and `delimiter` on the include line, and the alias is every row.

    file, err := mapper.NewDelimiterMapper("|", []string{"first_name", "last_name", "state"})
    if err != nil {
        fmt.Println(err)
        return
    }
    if err := file.LoadFile("people.csv"); err != nil {
        fmt.Println(err)
        return
    }
