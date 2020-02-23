package service

import (
	"fmt"
	"strings"

	"github.com/ozgio/strutil"

	"github.com/dave/jennifer/jen"
	"github.com/go-services/code"
	"github.com/go-services/source"
)

const (
	HTTPAnnotation = "http"
	GRPCAnnotation = "grpc"
)

type HTTPMethodRoute struct {
	Name       string
	Methods    []string
	MethodsAll string
	Route      string
}

type HTTPTransport struct {
	MethodRoutes []HTTPMethodRoute
}

func parseHTTPTransport(method source.InterfaceMethod) *HTTPTransport {
	httpAnnotations := source.FindAnnotations(HTTPAnnotation, &method)
	transport := &HTTPTransport{}
	if len(httpAnnotations) == 0 {
		return transport
	}
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

var typeFuncMap = map[string]struct {
	fn           string
	withoutError bool
}{
	"[]string": {
		fn:           "StringToStringArray",
		withoutError: true,
	},
	"int": {
		fn: "StringToInt",
	},
	"[]int": {
		fn: "StringToIntArray",
	},
	"float64": {
		fn: "StringToFloat64",
	},
	"[]float64": {
		fn: "StringToFloat64Array",
	},
	"float32": {
		fn: "StringToFloat32",
	},
	"[]float32": {
		fn: "StringToFloat32Array",
	},
	"bool": {
		fn: "StringToBool",
	},
}

func (h HTTPTransport) Decoder(ep Endpoint) string {
	if ep.Request == nil {
		return jen.Return(jen.Id("err")).GoString()
	}
	src := jen.Id("request").Op("=").Id(ep.Params[1].Type.String()).Block().Line()
	var vars []jen.Code
	var queries []jen.Code
	var body *jen.Statement
	for _, field := range ep.Request.Fields {
		if !isExported(field.Name) || field.Tags == nil {
			continue
		}
		if url := getTag("gos_url", *field.Tags); url != "" {
			tp := field.Type.String()
			if !isUrlTypeSupported(tp) {
				continue
			}
			if tp == "string" {
				vars = append(
					vars,
					jen.Id("request").Dot(field.Name).Op("=").Id("vars").Index(jen.Lit(url)).Line(),
				)
			} else {
				vars = append(vars, convertFunc(field.Name, url, tp, false))
			}
		} else if q := getTag("gos_query", *field.Tags); q != "" {
			tp := field.Type.String()
			if !isQueryTypeSupported(tp) {
				continue
			}
			if tp == "string" {
				queries = append(
					queries,
					jen.Id("request").Dot(field.Name).Op("=").Id(
						"r.URL.Query().Get",
					).Call(jen.Lit(q)).Line(),
				)
			} else {
				queries = append(queries, convertFunc(field.Name, q, tp, true))
			}
		} else if format := getTag("gos_body", *field.Tags); format != "" {
			parameterId := jen.Id("&request").Dot(field.Name)
			if field.Type.Pointer {
				parameterId = jen.Id("request").Dot(field.Name)
			}
			body = jen.Id("err").Op("=").Id("json").Dot("NewDecoder")
			body.Call(
				jen.Id("r.Body"),
			).Dot("Decode").Call(parameterId).Line()
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

func isUrlTypeSupported(tp string) bool {
	var supportedUrlTypes = []string{"string", "int", "float32", "float64"}
	found := false
	for _, supportedType := range supportedUrlTypes {
		if supportedType == tp {
			found = true
			break
		}
	}
	return found
}
func isQueryTypeSupported(tp string) bool {
	var supportedQueryTypes = []string{
		"string",
		"[]string",
		"int",
		"[]int",
		"bool",
		"float32",
		"[]float32",
		"float64",
		"[]float64",
	}
	found := false
	for _, supportedType := range supportedQueryTypes {
		if supportedType == tp {
			found = true
			break
		}
	}
	return found
}
func convertFunc(fieldName, varName, tp string, query bool) *jen.Statement {
	var value jen.Code
	tpFunc := typeFuncMap[tp]
	if query {
		value = jen.Id("r.URL.Query().Get").Call(jen.Lit(varName))
	} else {
		value = jen.Id("vars").Index(jen.Lit(varName))
	}
	operation := jen.Id("request").Dot(fieldName)
	if !tpFunc.withoutError {
		operation.Id(", err")
	}
	operation.Op("=").Id("utils").Dot(
		tpFunc.fn,
	).Call(value).Line()
	if !tpFunc.withoutError {
		operation = jen.Add(
			operation,
			jen.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jen.Id("request, errors.HTTPBadRequest(err.Error())")),
			).Line(),
		)
	}
	return operation
}
func getTag(key string, tags code.FieldTags) string {
	tag, _ := tags[key]
	return tag
}

var protoTypeMap = map[string]string{
	"float":   "double",
	"float64": "double",
	"float32": "float",
	"int32":   "int32",
	"int64":   "int64",
	"int":     "int64",
	"uint32":  "uint32",
	"uint64":  "uint64",
	"bool":    "bool",
	"string":  "string",
}

type ProtoMessage struct {
	Name   string
	Params []ProtoMessageParam
}
type ProtoMessageParam struct {
	Repeat   bool
	Name     string
	Type     string
	Position int
}
type GRPCEndpoint struct {
	messages []ProtoMessage
}

func (p *ProtoMessageParam) String() string {
	s := ""
	if p.Repeat {
		s += "repeated"
	}
	return fmt.Sprintf("%s %s %s = %d", s, p.Type, p.Name, p.Position)
}

func (m *ProtoMessage) String() string {
	s := fmt.Sprintf("message %s {\n", m.Name)
	for _, v := range m.Params {
		s += v.String() + "\n"
	}
	s += "}\n"
	return s
}

type GRPCTransport struct {
	endpoints []GRPCEndpoint
}

func parseGRPCTransport(svc *Service) *GRPCTransport {
	tp := &GRPCTransport{}
	structsCreated := map[string]string{}
	for i, ep := range svc.Endpoints {
		fmt.Println(ep.Name, structsCreated)
		tp.endpoints = append(tp.endpoints, parseGRPCEndpoint(ep, structsCreated))
		for _, msg := range tp.endpoints[i].messages {
			fmt.Println(msg.String())
		}
	}
	return tp
}
func parseGRPCEndpoint(ep Endpoint, structsCreated map[string]string) GRPCEndpoint {
	var messages []ProtoMessage
	if ep.Request != nil {
		generateMessage(&messages, ep.Name+"Request", ep.Request, structsCreated)
	}
	responseMessage := ProtoMessage{
		Name: ep.Name + "Response",
		Params: []ProtoMessageParam{
			{
				Repeat:   false,
				Name:     "Err",
				Type:     "string",
				Position: 1,
			},
		},
	}
	if ep.Response != nil {
		messageName := ep.Name + "BaseResponse"
		generateMessage(&messages, messageName, ep.Response, structsCreated)
		responseMessage.Params = append(responseMessage.Params, ProtoMessageParam{
			Repeat:   false,
			Name:     "Response",
			Type:     messageName,
			Position: 2,
		})

	}
	messages = append(messages, responseMessage)
	return GRPCEndpoint{
		messages: messages,
	}
}

func getStructPathName(imp *code.Import, name string) string {
	if imp == nil {
		return ""
	}
	return imp.FilePath + getStructName(imp, name)
}
func getStructName(imp *code.Import, name string) string {
	return strings.Title(strutil.ToCamelCase(imp.Alias)) + name
}

func generateMessage(messages *[]ProtoMessage, name string, structure *code.Struct, structsCreated map[string]string) {
	message := ProtoMessage{
		Name: name,
	}
	position := 1
	for _, field := range structure.Fields {
		tag := ""
		if field.Tags != nil {
			tag = getTag("gos_grpc", *field.Tags)
		}
		if !isExported(field.Name) && tag == "-" {
			continue
		}
		if field.Type.Import != nil {
			messageName, ok := structsCreated[getStructPathName(field.Type.Import, field.Type.Qualifier)]
			if !ok {
				s, err := findStruct(field.Type)
				if err != nil {
					panic(err)
				}
				messageName = name + "_" + getStructName(field.Type.Import, s.Name)
				generateMessage(messages, messageName, s, structsCreated)
			}
			structsCreated[getStructPathName(field.Type.Import, field.Type.Qualifier)] = messageName
			message.Params = append(message.Params, ProtoMessageParam{
				Repeat:   field.Type.ArrayType,
				Name:     field.Name,
				Type:     messageName,
				Position: position,
			})
		} else {
			if field.Type.String() == "[]byte" {
				message.Params = append(message.Params, ProtoMessageParam{
					Name:     field.Name,
					Type:     "bytes",
					Position: position,
				})
			} else if field.Type.MapType != nil {

			}
			tp, ok := protoTypeMap[field.Type.Qualifier]
			if !ok {
				fmt.Printf("Type %s not supported\n", tp)
				continue
			}
			message.Params = append(message.Params, ProtoMessageParam{
				Repeat:   field.Type.ArrayType,
				Name:     field.Name,
				Type:     tp,
				Position: position,
			})
		}
		position++
	}
	*messages = append(*messages, message)
}
