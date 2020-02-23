package generator

import (
	"errors"
	"fmt"
	"gos/fs"

	"gos/service"
	"log"

	"github.com/go-services/source"
)

func Generate(name, mod, httpAddress string) error {
	log.Printf("Generating service %s", name)
	data, err := fs.ReadFile(fs.AppFs(), fmt.Sprintf("%s/service.go", name))
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
