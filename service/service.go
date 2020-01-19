package service

import (
	"errors"
	"fmt"
	"gos/fs"
	"gos/template"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-services/code"
	"github.com/go-services/source"
	"github.com/ozgio/strutil"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	ANNOTATION = "service"
)

type EndpointMiddleware struct {
	Alias  string
	Method string
}

type Endpoint struct {
	Name string

	RootPkg              string
	ServiceInterfaceName string
	serviceFs            afero.Fs

	Params  []code.Parameter
	Results []code.Parameter

	Request  *code.Struct
	Response *code.Struct

	RequestImport  *code.Import
	ResponseImport *code.Import

	Middlewares        []EndpointMiddleware
	MiddlewarePackages map[string]string

	HTTPTransport     *HTTPTransport
	HTTPDecoderSource string
}

type Service struct {
	// interface name
	InterfaceName string
	HTTPAddress   string
	RootPkg       string
	ServiceName   string
	Package       string
	Module        string
	serviceFs     afero.Fs
	Endpoints     []Endpoint
}

func NewFromSource(src source.Source, svcName, mod, httpAddress string) (*Service, error) {
	inf := FindServiceInterface(src)
	if inf == nil {
		return nil, fmt.Errorf(
			"error while parsing service : %s",
			"Could not find service interface, make sure you are using @service()",
		)
	}
	svc := &Service{
		ServiceName:   svcName,
		RootPkg:       fmt.Sprintf("%s/%s", mod, svcName),
		InterfaceName: inf.Name(),
		HTTPAddress:   httpAddress,
		serviceFs:     afero.NewBasePathFs(fs.AppFs(), svcName),
		Package:       src.Package(),
		Module:        mod,
	}
	eps, err := svc.parseEndpoints(findServiceMethods(*inf))
	if err != nil {
		return nil, err
	}
	svc.Endpoints = eps
	return svc, nil
}

func (s Service) fixMethodImport(param code.Parameter) code.Parameter {
	if param.Type.Import == nil && isExported(param.Type.Qualifier) {
		currentPath, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		if viper.GetString("testPath") != "" {
			currentPath = path.Join(currentPath, viper.GetString("testPath"))
		}
		param.Type.Import = code.NewImportWithFilePath(
			"service",
			fmt.Sprintf("%s/%s", s.Module, s.Package),
			path.Join(currentPath, s.ServiceName),
		)
	}
	return param
}
func (s Service) parseEndpoints(methods []source.InterfaceMethod) (eps []Endpoint, err error) {
	for _, method := range methods {
		ep := Endpoint{
			RootPkg:              s.RootPkg,
			ServiceInterfaceName: s.InterfaceName,
			serviceFs:            s.serviceFs,
			Name:                 method.Name(),
		}
		if err := checkMethodParams(method.Params()); err != nil {
			return nil, err
		}
		if err := checkMethodResults(method.Results()); err != nil {
			return nil, err
		}
		for i, param := range method.Params() {
			fixedParam := s.fixMethodImport(param)
			ep.Params = append(ep.Params, fixedParam)
			if i == 1 {
				request, err := findStruct(fixedParam)
				if err != nil {
					return nil, err
				}
				ep.Request = request
				ep.RequestImport = fixedParam.Type.Import
			}
		}
		resultsLength := len(method.Results())
		for i, param := range method.Results() {
			fixedParam := s.fixMethodImport(param)
			ep.Results = append(ep.Results, fixedParam)
			if resultsLength > 1 && i == 0 {
				response, err := findStruct(fixedParam)
				if err != nil {
					return nil, err
				}
				ep.Response = response
				ep.ResponseImport = fixedParam.Type.Import
			}
		}
		ep.HTTPTransport = parseHTTPTransport(method)
		ep.HTTPDecoderSource = ep.HTTPTransport.Decoder(ep)
		ep.MiddlewarePackages, ep.Middlewares = parseMiddleware(method)
		eps = append(eps, ep)
	}
	return
}
func (s Service) Generate() error {
	err := fs.DeleteFolder(s.serviceFs, "gen")
	if err != nil {
		return err
	}
	files := map[string]string{
		"templates/service/gen/service.go.gotmpl":             "gen/gen.go",
		"templates/service/gen/cmd/cmd.go.gotmpl":             "gen/cmd/cmd.go",
		"templates/service/gen/errors/errors.go.gotmpl":       "gen/errors/errors.go",
		"templates/service/gen/errors/http.go.gotmpl":         "gen/errors/http.go",
		"templates/service/gen/utils/utils.go.gotmpl":         "gen/utils/utils.go",
		"templates/service/gen/endpoint/endpoint.go.gotmpl":   "gen/endpoint/endpoint.go",
		"templates/service/gen/transport/transport.go.gotmpl": "gen/transport/transport.go",
		"templates/service/gen/transport/http/http.go.gotmpl": "gen/transport/http/http.go",
	}
	for k, v := range files {
		if err := generateFile(s.serviceFs, k, v, s); err != nil {
			return err
		}
	}
	if err := s.generateEndpoints(); err != nil {
		return err
	}
	if err := s.generateCmd(); err != nil {
		return err
	}
	return nil
}

func (s Service) generateEndpoints() error {
	for _, ep := range s.Endpoints {
		if err := ep.Generate(); err != nil {
			return err
		}
	}
	return nil
}

func (s Service) generateCmd() error {
	if b, err := afero.Exists(s.serviceFs, "cmd/main.go"); err != nil {
		return err
	} else if b {
		return nil
	}
	files := map[string]string{
		"templates/service/cmd/main.go.gotmpl": "cmd/main.go",
	}
	for k, v := range files {
		if err := generateFile(s.serviceFs, k, v, s); err != nil {
			return err
		}
	}
	return nil
}

func (e Endpoint) Generate() error {
	files := map[string]string{
		"templates/service/gen/endpoint/definitions/method.go.gotmpl": fmt.Sprintf("gen/endpoint/definitions/%s.go", strutil.ToSnakeCase(e.Name)),
		"templates/service/gen/endpoint/method.go.gotmpl":             fmt.Sprintf("gen/endpoint/%s.go", strutil.ToSnakeCase(e.Name)),
		"templates/service/gen/transport/http/method.go.gotmpl":       fmt.Sprintf("gen/transport/http/%s.go", strutil.ToSnakeCase(e.Name)),
	}
	for k, v := range files {
		if err := generateFile(e.serviceFs, k, v, e); err != nil {
			return err
		}
	}
	return nil
}

func findServiceMethods(inf source.Interface) (methods []source.InterfaceMethod) {
	for _, method := range inf.Methods() {
		if isExported(method.Name()) {
			methods = append(methods, method)
		}
	}
	return methods
}
func FindServiceInterface(src source.Source) *source.Interface {
	for _, inf := range src.Interfaces() {
		annotations := source.FindAnnotations(ANNOTATION, &inf)
		if len(annotations) > 0 {
			return &inf
		}
	}
	return nil
}
func findStruct(param code.Parameter) (*code.Struct, error) {
	notFoundErr := errors.New(
		"could not find structure, make sure that you are using a structure as request/response parameters",
	)
	if param.Type.Import.FilePath == "" {
		return nil, notFoundErr
	}
	fls, err := ioutil.ReadDir(param.Type.Import.FilePath)
	if err != nil {
		panic(err)
	}
	if fls == nil {
		return nil, notFoundErr
	}
	for _, file := range fls {
		if file.IsDir() {
			continue
		}
		data, err := ioutil.ReadFile(path.Join(param.Type.Import.FilePath, file.Name()))
		if err != nil {
			return nil, err
		}
		dataFile, err := source.New(string(data))
		if err != nil {
			return nil, err
		}
		for _, structure := range dataFile.Structures() {
			if structure.Name() == param.Type.Qualifier {
				return structure.Code().(*code.Struct), nil
			}
		}
	}
	return nil, notFoundErr
}
func isExported(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(ch)
}

func checkMethodParams(params []code.Parameter) error {
	if len(params) != 1 && len(params) != 2 {
		return errors.New("method must except either the context or the context and the request struct")
	}
	if !(params[0].Type.Qualifier == "Context" &&
		params[0].Type.Import.Path == "context") &&
		params[0].Type.Pointer &&
		params[0].Type.Variadic {
		return errors.New("the first parameter of the method needs to be the context")
	}
	if len(params) == 2 && !isExported(params[1].Type.Qualifier) {
		return errors.New("request needs to be an exported structure")
	}
	return nil
}

func checkMethodResults(params []code.Parameter) error {
	if (len(params) != 1 && len(params) != 2) ||
		len(params) == 1 && params[0].Type.Qualifier != "error" ||
		len(params) == 2 && params[1].Type.Qualifier != "error" ||
		len(params) == 2 && !params[0].Type.Pointer ||
		len(params) == 2 && !isExported(params[0].Type.Qualifier) {
		return errors.New("method must return either the error or the response pointer and the error")
	}
	return nil
}
func generateFile(serviceFs afero.Fs, tpl, path string, data interface{}) error {
	src, err := template.CompileGoFromPath(tpl, data)
	if err != nil {
		return err
	}
	return fs.WriteFile(serviceFs, path, src)
}

func parseMiddleware(method source.InterfaceMethod) (packages map[string]string, mdw []EndpointMiddleware) {
	annotations := source.FindAnnotations("middleware", &method)
	packages = map[string]string{}

	for _, v := range annotations {
		pth := v.Get("path").String()
		pathParts := strings.Split(pth, ".")
		if len(pathParts) == 1 {
			mdw = append(mdw, EndpointMiddleware{
				Alias:  "",
				Method: pathParts[0],
			})
			continue
		}
		ep := EndpointMiddleware{
			Alias:  "",
			Method: pathParts[len(pathParts)-1],
		}
		pkg := strings.Join(pathParts[:len(pathParts)-1], "/")
		if v, ok := packages[pkg]; ok {
			ep.Alias = v
		} else {
			packages[pkg] = fmt.Sprintf("mdw%d", len(packages)+1)
			ep.Alias = packages[pkg]
		}
		mdw = append(mdw, ep)
	}
	return
}
