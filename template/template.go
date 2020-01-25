package template

import (
	"bytes"
	"gos/fs"
	"io/ioutil"
	"text/template"

	"github.com/spf13/afero"

	"golang.org/x/tools/imports"
)

func GenerateFile(serviceFs afero.Fs, tpl, path string, data interface{}) error {
	src, err := CompileGoFromPath(tpl, data)
	if err != nil {
		return err
	}
	return fs.WriteFile(serviceFs, path, src)
}

func CompileFromPath(tplPath string, data interface{}) (string, error) {
	buf, err := FromPath(tplPath)
	if err != nil {
		return "", err
	}
	t := template.Must(template.New(tplPath).Funcs(CustomFunctions).Parse(buf))
	templateBuffer := bytes.NewBufferString("")
	err = t.Execute(templateBuffer, data)
	if err != nil {
		return "", err
	}
	return templateBuffer.String(), err
}
func CompileGoFromPath(tplPath string, data interface{}) (string, error) {
	src, err := CompileFromPath(tplPath, data)
	if err != nil {
		return "", err
	}
	prettyCode, err := imports.Process("template.go", []byte(src), nil)
	return string(prettyCode), err
}

func FromPath(tplPath string) (string, error) {
	file, err := FS.Open("/assets/" + tplPath)
	if err != nil {
		return "", err
	}
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
