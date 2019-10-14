package service

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/go-services/code"
	"github.com/go-services/source"
)

const (
	HTTPAnnotation = "http"
)

type HTTPMethodRoute struct {
	Name       string
	Request    code.Struct
	Methods    []string
	MethodsAll string
	Route      string
}

type HTTPTransport struct {
	MethodRoutes []HTTPMethodRoute
}

func parseHTTPTransport(method source.InterfaceMethod) *HTTPTransport {
	httpAnnotations := source.FindAnnotations(HTTPAnnotation, &method)
	if len(httpAnnotations) == 0 {
		return nil
	}
	transport := &HTTPTransport{}
	for _, httpAnnotation := range httpAnnotations {
		var methodsPrepared []string
		for _, method := range strings.Split(httpAnnotation.Get("methods").String(), ",") {
			methodsPrepared = append(methodsPrepared, strings.ToUpper(strings.TrimSpace(method)))
		}
		route := httpAnnotation.Get("route").String()
		if !strings.HasPrefix(route, "/") {
			route = "/" + route
		}
		transport.MethodRoutes = append(
			transport.MethodRoutes,
			HTTPMethodRoute{
				Name:       httpAnnotation.Get("name").String(),
				Methods:    methodsPrepared,
				MethodsAll: strings.Join(methodsPrepared, ", "),
				Route:      route,
			},
		)
	}
	return transport
}

func (h HTTPTransport) Decoder(ep Endpoint) string {
	if ep.Request == nil {
		return jen.Return(jen.Id("err")).GoString()
	}
	src := jen.Id("request").Op("=").Id(ep.Params[1].Type.String()).Block().Line()
	var vars []jen.Code
	var queries []jen.Code
	var body jen.Code
	for _, field := range ep.Request.Fields {
		if !isExported(field.Name) {
			continue
		}
		if url := getTag("gos_url", *field.Tags); url != "" {
			switch field.Type.String() {
			case "string":
				vars = append(
					vars,
					jen.Id("request").Dot(field.Name).Op("=").Id("vars").Index(jen.Lit(url)).Line(),
				)
			case "int":
				vars = append(
					vars,
					jen.Id("request").Dot(field.Name).Id(", err").Op("=").Id("utils").Dot(
						"StringToInt",
					).Call(jen.Id("vars").Index(jen.Lit(url))).Line(),
					jen.If(jen.Err().Op("!=").Nil()).Block(
						jen.Return(jen.Id("request, err")),
					).Line(),
				)
			case "float64":
				vars = append(
					vars,
					jen.Id("request").Dot(field.Name).Id(", err").Op("=").Id("utils").Dot(
						"StringToFloat64",
					).Call(jen.Id("vars").Index(jen.Lit(url))).Line(),
					jen.If(jen.Err().Op("!=").Nil()).Block(
						jen.Return(jen.Id("request, err")),
					).Line(),
				)
			case "float32":
				vars = append(
					vars,
					jen.Id("request").Dot(field.Name).Id(", err").Op("=").Id("utils").Dot(
						"StringToFloat32",
					).Call(jen.Id("vars").Index(jen.Lit(url))).Line(),
					jen.If(jen.Err().Op("!=").Nil()).Block(
						jen.Return(jen.Id("request, err")),
					).Line(),
				)
			}
		}
		if q := getTag("gos_query", *field.Tags); q != "" {
			switch field.Type.String() {
			case "string":
				queries = append(
					queries, //	r.URL.Query().Get()
					jen.Id("request").Dot(field.Name).Op("=").Id(
						"r.URL.Query().Get",
					).Call(jen.Lit(q)).Line(),
				)
			case "[]string":
				queries = append(
					queries,
					jen.Id("request").Dot(field.Name).Op("=").Id("utils").Dot(
						"StringToStringArray",
					).Call(jen.Id("r.URL.Query().Get").Call(jen.Lit(q))).Line(),
				)
			case "int":
				queries = append(
					queries,
					jen.Id("request").Dot(field.Name).Id(", err").Op("=").Id("utils").Dot(
						"StringToInt",
					).Call(jen.Id("r.URL.Query().Get").Call(jen.Lit(q))).Line(),
					jen.If(jen.Err().Op("!=").Nil()).Block(
						jen.Return(jen.Id("request, err")),
					).Line(),
				)
			case "[]int":
				queries = append(
					queries,
					jen.Id("request").Dot(field.Name).Id(", err").Op("=").Id("utils").Dot(
						"StringToIntArray",
					).Call(jen.Id("r.URL.Query().Get").Call(jen.Lit(q))).Line(),
					jen.If(jen.Err().Op("!=").Nil()).Block(
						jen.Return(jen.Id("request, err")),
					).Line(),
				)
			case "float64":
				queries = append(
					queries,
					jen.Id("request").Dot(field.Name).Id(", err").Op("=").Id("utils").Dot(
						"StringToFloat64",
					).Call(jen.Id("r.URL.Query().Get").Call(jen.Lit(q))).Line(),
					jen.If(jen.Err().Op("!=").Nil()).Block(
						jen.Return(jen.Id("request, err")),
					).Line(),
				)
			case "[]float64":
				queries = append(
					queries,
					jen.Id("request").Dot(field.Name).Id(", err").Op("=").Id("utils").Dot(
						"StringToFloat64Array",
					).Call(jen.Id("r.URL.Query().Get").Call(jen.Lit(q))).Line(),
					jen.If(jen.Err().Op("!=").Nil()).Block(
						jen.Return(jen.Id("request, err")),
					).Line(),
				)
			case "float32":
				queries = append(
					queries,
					jen.Id("request").Dot(field.Name).Id(", err").Op("=").Id("utils").Dot(
						"StringToFloat32",
					).Call(jen.Id("r.URL.Query().Get").Call(jen.Lit(q))).Line(),
					jen.If(jen.Err().Op("!=").Nil()).Block(
						jen.Return(jen.Id("request, err")),
					).Line(),
				)
			case "[]float32":
				queries = append(
					queries,
					jen.Id("request").Dot(field.Name).Id(", err").Op("=").Id("utils").Dot(
						"StringToFloat32Array",
					).Call(jen.Id("r.URL.Query().Get").Call(jen.Lit(q))).Line(),
					jen.If(jen.Err().Op("!=").Nil()).Block(
						jen.Return(jen.Id("request, err")),
					).Line(),
				)
			case "bool":
				queries = append(
					queries,
					jen.Id("request").Dot(field.Name).Id(", err").Op("=").Id("utils").Dot(
						"StringToBool",
					).Call(jen.Id("r.URL.Query().Get").Call(jen.Lit(q))).Line(),
					jen.If(jen.Err().Op("!=").Nil()).Block(
						jen.Return(jen.Id("request, err")),
					).Line(),
				)
			}
		}
		if format := getTag("gos_body", *field.Tags); format != "" {
			body = jen.Id("err").Op("=").Id("json").Dot("NewDecoder").Call(
				jen.Id("r.Body"),
			).Dot("Decode").Call(jen.Id("&request").Dot(field.Name)).Line()
		}
	}
	stmt := jen.Empty()
	if len(vars) > 0 {
		stmt.Add(jen.Id("vars").Op(":=").Id("mux").Dot("Vars").Call(jen.Id("r")).Line())
		stmt.Add(src)
		stmt.Add(vars...)
	}
	if len(queries) > 0 {
		stmt.Add(queries...)
	}
	if len(vars) == 0 && len(queries) == 0 && body == nil {
		body = jen.Id("err").Op("=").Id("json").Dot("NewDecoder").Call(
			jen.Id("r.Body"),
		).Dot("Decode").Call(jen.Id("&request")).Line()
	}
	if body != nil {
		stmt.Add(body)
	}
	stmt.Add(jen.Return(jen.Id("request, err")))
	return code.NewRawCode(stmt).String()
}

func getTag(key string, tags code.FieldTags) string {
	tag, _ := tags[key]
	return tag
}
