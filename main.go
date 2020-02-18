// Copyright (c) 2020 Kien Nguyen-Tuan <kiennt2609@gmail.com>
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

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

var readmeRaw string = `
1. [{{ .Title }}]({{ .File }})
`

type Page struct {
	Title string
	File  string
}

func NewPage(name string) Page {
	slug := strings.ToLower(strings.Join(strings.Fields(name), "-"))
	file := slug + ".md"
	return Page{
		Title: name,
		File:  file,
	}
}

func main() {
	flags := struct {
		base  string
		tmpl  string
		title string
	}{}

	cur, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	a := kingpin.New(filepath.Base(os.Args[0]), "A simple Golang document generator")
	a.HelpFlag.Short('h')
	a.Flag("base-dir", "The base directory stores all generated documentation.").
		Default(cur).StringVar(&flags.base)
	a.Flag("template", "The documentation template file.").
		StringVar(&flags.tmpl)
	a.Flag("title", "The full document title, e.g.: 'Using go templates guideline'").
		StringVar(&flags.title)
	_, err = a.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing commandline arguments"))
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	var (
		page       Page
		readmef    *os.File
		pagef      *os.File
		pageTmpl   *template.Template
		readmeTmpl *template.Template
		tmpRMLine  bytes.Buffer
	)

	defer func() {
		readmef.Close()
		pagef.Close()
	}()

	// If the file doesn't exist, create it, or append to the file
	readmef, err = os.OpenFile(filepath.Join(flags.base, "README.md"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	// Create page with template
	page = NewPage(flags.title)
	// Open file
	pagef, err = os.OpenFile(filepath.Join(flags.base, page.File),
		os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	// Create a template and parse the page into it.
	pageTmpl = template.Must(template.ParseFiles(flags.tmpl))
	if err = pageTmpl.Execute(pagef, page); err != nil {
		panic(err)
	}
	// Append line to README for index purpose
	readmeTmpl = template.Must(template.New("readme").Parse(readmeRaw))
	if err := readmeTmpl.Execute(&tmpRMLine, page); err != nil {
		panic(err)
	}
	// Check generated line is already existed in README or not.
	b, err := ioutil.ReadFile(filepath.Join(flags.base, "README.md"))
	if err != nil {
		panic(err)
	}
	if !strings.Contains(string(b), string(tmpRMLine.Bytes())) {
		if err := readmeTmpl.Execute(readmef, page); err != nil {
			panic(err)
		}
	}
}
