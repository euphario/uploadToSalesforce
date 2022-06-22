package handlers

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func ParseTemplates() *template.Template {
	templ := template.New("")
	err := filepath.Walk("./frontend/views", func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".html") {
			_, err = templ.ParseFiles(path)
			if err != nil {
				log.Println(err)
			}
		}

		return err
	})

	if err != nil {
		panic(err)
	}

	return templ
}
