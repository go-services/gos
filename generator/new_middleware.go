package generator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gos/config"
	"gos/fs"
	"gos/service"
	"gos/template"
	"strings"

	"github.com/go-services/source"

	"github.com/ozgio/strutil"

	"github.com/spf13/afero"
)

func NewMiddleware(name string, service string, endpoint string) error {
	appFs := fs.AppFs()
	b, err := afero.Exists(appFs, "kit.json")

	data, err := fs.ReadFile(appFs, fmt.Sprintf("%s/service.go", service))
	if err != nil {
		return errors.New("A read error occurred. Please update your code..: " + err.Error())
	}
	name = strings.Title(strutil.ToCamelCase(name))
	if err != nil {
		return err
	} else if !b {
		return errors.New("not in a kit project, you need to be in a project to run this command")
	}

	configData, err := fs.ReadFile(appFs, "kit.json")
	if err != nil {
		return errors.New("could not read kit.json")
	}
	var kitConfig config.KitConfig
	err = json.NewDecoder(bytes.NewBufferString(configData)).Decode(&kitConfig)
	if err != nil {
		return err
	}

	middlewareFilePath := fmt.Sprintf("%s/middleware/%s.go", service, endpoint)
	b, err = afero.Exists(appFs, middlewareFilePath)
	var fileData string
	if !b {
		fileData = createEndpointMiddlewareFile(kitConfig.Module, service)
	} else {
		fileData, err = fs.ReadFile(appFs, middlewareFilePath)
		if err != nil {
			return err
		}
	}

	data := map[string]string{
		"Name":         name,
		"EndpointName": strings.Title(strutil.ToCamelCase(endpoint)),
	}
	mdwMethod, err := template.CompileFromPath("templates/partials/endpoint_middleware.go.gotmpl", data)
	if err != nil {
		return err
	}
	fileData += mdwMethod
	if err := fs.WriteFile(appFs, middlewareFilePath, fileData); err != nil {
		return err
	}
	return annotateEndpoint(appFs, kitConfig, service, endpoint, name)

}

func annotateEndpoint(appFs afero.Fs, kitConfig config.KitConfig, svc string, endpoint string, name string) error {
	data, err := fs.ReadFile(appFs, fmt.Sprintf("%s/service.go", svc))
	if err != nil {
		return errors.New("A read error occurred. Please update your code..: " + err.Error())
	}
	src, err := source.New(data)
	if err != nil {
		return errors.New("A read error occurred. Please update your code..: " + err.Error())
	}
	inf := service.FindServiceInterface(*src)

	if inf == nil {
		return fmt.Errorf(
			"error while parsing service : %s",
			"Could not find service interface, make sure you are using @service()",
		)
	}
	ep := strings.Title(strutil.ToCamelCase(endpoint))
	mdwPath := fmt.Sprintf("%s.%s.%s.%s", kitConfig.Module, svc, "middleware", name)
	err = src.CommentInterfaceMethod(inf.Name(), ep, fmt.Sprintf("@middleware(path=\"%s\")", mdwPath))
	if err != nil {
		return err
	}
	newService, err := src.String()
	if err != nil {
		return err
	}
	return fs.WriteFile(appFs, fmt.Sprintf("%s/service.go", svc), newService)
}

func createEndpointMiddlewareFile(mod, service string) string {
	return fmt.Sprintf(`package middleware

import (
	"%s/%s"
	"%s/%s/gen/endpoint/definitions"
	"context"
)`, mod, service, mod, service)
}
