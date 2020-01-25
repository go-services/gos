package generator

import (
	"errors"
	"fmt"
	"gos/fs"

	"github.com/go-services/source"
	"github.com/spf13/afero"

	"gos/service"
	"log"
)

func Generate(name, mod, httpAddress string, rootFs afero.Fs) error {
	log.Printf("Generating service %s", name)
	data, err := fs.ReadFile(rootFs, fmt.Sprintf("%s/service.go", name))
	if err != nil {
		return errors.New("A read error occurred. Please update your code..: " + err.Error())
	}
	src, err := source.New(data)
	if err != nil {
		return errors.New("A read error occurred. Please update your code..: " + err.Error())
	}
	srv, err := service.NewFromSource(*src, name, mod, httpAddress)
	if err != nil {
		return errors.New("A read error occurred. Please update your code..: " + err.Error())
	}
	if err := srv.Generate(); err != nil {
		return errors.New("A read error occurred. Please update your code..: " + err.Error())
	}
	return nil
}
