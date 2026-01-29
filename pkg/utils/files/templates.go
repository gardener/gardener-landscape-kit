// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package files

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"text/template"

	"github.com/go-sprout/sprout/sprigin"
)

// RenderTemplateFiles renders all template files in the given templateDir from the provided embed.FS using the provided vars.
func RenderTemplateFiles(templateFS embed.FS, templateDir string, vars map[string]any) (map[string][]byte, error) {
	return renderFilesInDir(templateFS, templateDir, "", vars)
}

func renderFilesInDir(templateFS embed.FS, templateDir, currentDir string, vars map[string]any) (map[string][]byte, error) {
	var objects = make(map[string][]byte)

	dir, err := templateFS.ReadDir(path.Join(templateDir, currentDir))
	if err != nil {
		return nil, err
	}
	for _, file := range dir {
		fileName := file.Name()
		if file.IsDir() {
			subDir := path.Join(currentDir, fileName)
			subDirObjects, err := renderFilesInDir(templateFS, templateDir, subDir, vars)
			if err != nil {
				return nil, err
			}
			for fileName, fileContents := range subDirObjects {
				objects[path.Join(subDir, fileName)] = fileContents
			}
			continue
		}
		fileContents, err := templateFS.ReadFile(path.Join(templateDir, currentDir, fileName))
		if err != nil {
			return nil, err
		}

		if fileContents, fileName, err = RenderTemplate(fileContents, fileName, vars); err != nil {
			return nil, err
		}
		objects[fileName] = fileContents
	}

	return objects, nil
}

// RenderTemplate renders the given fileContents as a Go text/template using the provided vars.
func RenderTemplate(fileContents []byte, fileName string, vars map[string]any) ([]byte, string, error) {
	fileTemplate, err := template.New(fileName).Funcs(sprigin.TxtFuncMap()).Parse(string(fileContents))
	if err != nil {
		return nil, "", fmt.Errorf("error parsing template '%s': %w", fileName, err)
	}

	var templatedResult bytes.Buffer
	if err := fileTemplate.Execute(&templatedResult, vars); err != nil {
		return nil, "", fmt.Errorf("error executing '%s' template: %w", fileName, err)
	}

	return templatedResult.Bytes(), fileName, nil
}
