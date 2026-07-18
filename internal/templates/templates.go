package templates

import (
	"bytes"
	"embed"
	"path/filepath"
	"strings"
	"text/template"
)

var tmpl *template.Template

//go:embed files/*.txt
var templateFolder embed.FS

const templateFolderName = "files"

func init() {
	tmpl = template.New("root")

	entries, err := templateFolder.ReadDir(templateFolderName)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()

		templateName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		fileContent, err := templateFolder.ReadFile(templateFolderName + "/" + fileName)
		if err != nil {
			panic(err)
		}

		_, err = tmpl.New(templateName).Parse(string(fileContent))
		if err != nil {
			panic(err)
		}
	}
}

func Render(name string, data any) string {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "ErrTemplateInvalid"
	}
	return buf.String()
}
