package main

import (
	"embed"
	"path"
	"sort"
	"strings"
)

// The examples of limu's gallery (internal/gallery/examples in limu's
// repository, under its MIT license), a file for each type of diagram: front
// matter with the type's title, then each example as a heading and a Mermaid
// code block.
//
//go:embed samples/*.md
var sampleFiles embed.FS

// A sampleType is a type of diagram and its examples.
type sampleType struct {
	title    string
	examples []sample
}

// A sample is an example's title and its Mermaid source.
type sample struct {
	title, src string
}

// samples are the types of diagram with their examples, by title.
func samples() []sampleType {
	names, err := sampleFiles.ReadDir("samples")
	if err != nil {
		panic(err)
	}
	var types []sampleType
	for _, n := range names {
		data, err := sampleFiles.ReadFile(path.Join("samples", n.Name()))
		if err != nil {
			panic(err)
		}
		types = append(types, parseSamples(string(data)))
	}
	sort.Slice(types, func(i, j int) bool { return types[i].title < types[j].title })
	return types
}

// parseSamples reads a file of the gallery: the title in its front matter,
// and each heading with the code block under it.
func parseSamples(text string) sampleType {
	var t sampleType
	lines := strings.Split(text, "\n")
	front := len(lines) > 0 && lines[0] == "---"
	var title string
	var code []string
	inCode := false
	for i, line := range lines {
		switch {
		case front:
			if i > 0 && line == "---" {
				front = false
			} else if v, ok := strings.CutPrefix(line, "title: "); ok {
				t.title = v
			}
		case inCode:
			if line == "```" {
				t.examples = append(t.examples, sample{title, strings.Join(code, "\n")})
				inCode, code = false, nil
			} else {
				code = append(code, line)
			}
		case line == "```mermaid":
			inCode = true
		default:
			if v, ok := strings.CutPrefix(line, "## "); ok {
				title = v
			}
		}
	}
	return t
}
