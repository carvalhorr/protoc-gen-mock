package main

import (
	"flag"
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"strconv"
	"strings"
)

const (
	fmtPackage         = protogen.GoImportPath("fmt")
	osPackage          = protogen.GoImportPath("os")
	strConvPackage     = protogen.GoImportPath("strconv")
	contextPackage     = protogen.GoImportPath("context")
	protoPackage       = protogen.GoImportPath("github.com/golang/protobuf/proto")
	grpcPackage        = protogen.GoImportPath("google.golang.org/grpc")
	grpcHandlerPackage = protogen.GoImportPath("github.com/thebaasco/protoc-gen-mock/grpchandler")
	stubPackage        = protogen.GoImportPath("github.com/thebaasco/protoc-gen-mock/stub")
	remotePackage      = protogen.GoImportPath("github.com/thebaasco/protoc-gen-mock/remote")
	codesPackage       = protogen.GoImportPath("google.golang.org/grpc/codes")
	statusPackage      = protogen.GoImportPath("google.golang.org/grpc/status")
	bootstrapPackage   = protogen.GoImportPath("github.com/thebaasco/protoc-gen-mock/bootstrap")
	deprecationComment = "// Deprecated: Do not use."
)

func main() {
	var (
		flags flag.FlagSet
		//plugins      = flags.String("plugins", "", "list of plugins to enable (supported values: grpc)")
		importPrefix = flags.String("import_prefix", "", "prefix to prepend to import paths")
	)
	importRewriteFunc := func(importPath protogen.GoImportPath) protogen.GoImportPath {
		switch importPath {
		case "context", "fmt", "math":
			return importPath
		}
		if *importPrefix != "" {
			return protogen.GoImportPath(*importPrefix) + importPath
		}
		return importPath
	}
	protogen.Options{
		ParamFunc:         flags.Set,
		ImportRewriteFunc: importRewriteFunc,
	}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			GenerateFile(gen, f)
		}
		GenerateMain(gen)
		GenerateDockerfile(gen)
		return nil
	})
}

func GenerateMain(gen *protogen.Plugin) {
	if len(gen.Files) == 0 {
		return
	}
	packageName := gen.Files[0].GoPackageName
	file := gen.NewGeneratedFile("main.go", "")
	file.P("package ", packageName)
	file.P("")
	file.P("func main() {")
	file.P("restPort, found := ", osPackage.Ident("LookupEnv"), "(\"REST_PORT\")")
	file.P("if !found {")
	file.P("restPort = \"1068\" // default REST port")
	file.P("}")
	file.P("restP, err := ", strConvPackage.Ident("ParseUint"), "(restPort, 10, 64)")
	file.P("if err != nil {")
	file.P("restP = 1068")
	file.P("}")
	file.P("grpcPort, found := ", osPackage.Ident("LookupEnv"), "(\"GRPC_PORT\")")
	file.P("if !found {")
	file.P("grpcPort = \"10010\" // default GRPC port")
	file.P("}")
	file.P("grpcP, err := ", strConvPackage.Ident("ParseUint"), "(grpcPort, 10, 64)")
	file.P("if err != nil {")
	file.P("grpcP = 10010")
	file.P("}")
	file.P("Start(uint(restP), uint(grpcP), \"./tmp\")")
	file.P("}")
	file.P("")
	file.P("func Start(restPort, grpcPort uint, tmpPath string) {")
	file.P(bootstrapPackage.Ident("BootstrapServers"), "(tmpPath, restPort, grpcPort, MockServicesRegistrationCallback)")
	file.P("}")
	file.P("var MockServicesRegistrationCallback = func(stubsMatcher ", stubPackage.Ident("StubsMatcher"), ") ", grpcHandlerPackage.Ident("MockService"), " {")
	file.P("return ", grpcHandlerPackage.Ident("NewCompositeMockService"), "([]", grpcHandlerPackage.Ident("MockService"), "{")
	for _, f := range gen.Files {
		for _, s := range f.Services {
			file.P("New", s.GoName, "MockService(stubsMatcher),")
		}
	}
	file.P("})")
	file.P("}")

}

func GenerateDockerfile(gen *protogen.Plugin) {
	if len(gen.Files) == 0 {
		return
	}
	packageName := gen.Files[0].GoPackageName
	file := gen.NewGeneratedFile("Dockerfile", "")
	file.P("FROM golang:1.15.8-alpine3.13 as builder")
	file.P("RUN apk add build-base")
	file.P("WORKDIR /mock")
	file.P("COPY *.go ./")
	file.P("RUN sed -i 's/package ", packageName, "/package main/' *.go")
	file.P("RUN go mod init ", packageName)
	file.P("RUN go build -o app")
	file.P("FROM golang:1.15.8-alpine3.13")
	file.P("COPY  --from=builder /mock/app .")
	file.P("ENTRYPOINT ./app")
}

// GenerateFile generates a .mock.pb.go file containing gRPC service definitions.
func GenerateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if !file.Generate {
		return nil
	}
	if len(file.Services) == 0 {
		return nil
	}
	// fmt.Println("FILENAME ", file.GeneratedFilenamePrefix)
	filename := file.GeneratedFilenamePrefix + ".mock.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	mockGenerator := mockServicesGenerator{
		gen:  gen,
		file: file,
		g:    g,
	}
	mockGenerator.genHeader(string(file.GoPackageName))
	mockGenerator.GenerateFileContent()
	return g
}

type mockServicesGenerator struct {
	gen  *protogen.Plugin
	file *protogen.File
	g    *protogen.GeneratedFile
}

// GenerateFileContent generates the gRPC service definitions, excluding the package statement.
func (m mockServicesGenerator) GenerateFileContent() {
	if len(m.file.Services) == 0 {
		return
	}
	for _, service := range m.file.Services {
		m.genService(service)
	}
}

func (m mockServicesGenerator) genService(service *protogen.Service) {
	m.genMockServiceConstructor(service)
	m.genMockServiceDefinition(service)
	m.genMockServiceRegistrationFunction(service)
	m.genGetSupportedMethodsFunction(service)
	m.genGetPayloadExamplesFunction(service)
	m.genGetRequestInstance(service)
	m.genGetResponseInstance(service)
	m.genGetStubsValidator(service)
	m.genIsValid(service)
	m.genForwardRequest(service)
	m.genRemoteClient(service)
	m.genMockServiceDescriptor(service)
	for _, method := range service.Methods {
		methodHandlerName := m.getMethodHandlerName(service, method)
		m.genMockMethodHandler(service, method, methodHandlerName)
	}
}

func (m mockServicesGenerator) genHeader(packageName string) {
	m.g.P("// Code generated by protoc-gen-mock. DO NOT EDIT.")
	m.g.P()
	m.g.P("package ", packageName)
	m.g.P()
}

func (m mockServicesGenerator) genMockServiceConstructor(service *protogen.Service) {
	serviceName := m.getMockServiceName(service)
	m.g.P("func New", serviceName, "(stubsMatcher ", stubPackage.Ident("StubsMatcher"), ") ", grpcHandlerPackage.Ident("MockService"), "{")
	m.g.P("return &", unexport(serviceName), "{")
	m.g.P("StubsMatcher: stubsMatcher,")
	m.g.P("}")
	m.g.P("}")
}

func (m mockServicesGenerator) genMockServiceDefinition(service *protogen.Service) {
	serviceName := m.getMockServiceName(service)
	serviceBaseInterfaceName := m.getMockServerBaseInterfaceName(service)
	m.g.P("// ", serviceName, " is the server API for ", service.GoName, " service.")
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		m.g.P("//")
		m.g.P(deprecationComment)
	}
	m.g.Annotate(serviceName, service.Location)
	m.g.P("type ", unexport(serviceName), " struct {")
	m.g.P(serviceBaseInterfaceName)
	m.g.P("StubsMatcher ", stubPackage.Ident("StubsMatcher"))
	m.g.P("}")
	m.g.P()
}

func (m mockServicesGenerator) genMockServiceRegistrationFunction(service *protogen.Service) {
	// Server registration.
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		m.g.P(deprecationComment)
	}
	m.g.P("func (mock *", unexport(m.getMockServiceName(service)), ") Register(s *", grpcPackage.Ident("Server"), ") {")
	m.g.P("s.RegisterService(&", m.getMockServiceDescriptorName(service), ", ", m.getMockServerBaseInterfaceName(service), "(mock))")
	m.g.P("}")
	m.g.P()
}

func (m mockServicesGenerator) genGetSupportedMethodsFunction(service *protogen.Service) {
	m.g.P("func (mock *", unexport(m.getMockServiceName(service)), ") GetSupportedMethods() []string {")
	m.g.P("return []string{")
	for _, method := range service.Methods {
		m.g.P(m.getFullMethodName(service, method), ",")
	}
	m.g.P("}")
	m.g.P("}")
	m.g.P()
}

func (m mockServicesGenerator) genGetPayloadExamplesFunction(service *protogen.Service) {
	m.g.P("func (mock *", unexport(m.getMockServiceName(service)), ") GetPayloadExamples() []", stubPackage.Ident("Stub"), "{")
	m.g.P("return []", stubPackage.Ident("Stub"), "{")
	for _, method := range service.Methods {
		m.g.P("{")
		m.g.P("FullMethod: ", m.getFullMethodName(service, method), ",")
		m.g.P("Type: ", strconv.Quote("mock | forward"), ",")
		m.g.P("Request: &", stubPackage.Ident("StubRequest"), " {")
		m.g.P("Match: \"exact | partial\",")
		m.g.P("Content: ", stubPackage.Ident("JsonString"), "(", stubPackage.Ident("CreateStubExample"), "(new(", method.Input.GoIdent, "))", "),")
		m.g.P("Metadata: make(map[string][]string, 0),")
		m.g.P("},")
		m.g.P("Response: &", stubPackage.Ident("StubResponse"), " {")
		m.g.P("Type: ", strconv.Quote("success | error"), ", ")
		m.g.P("Content: ", stubPackage.Ident("JsonString"), "(", stubPackage.Ident("CreateStubExample"), "(new(", method.Output.GoIdent, "))", "),")
		m.g.P("},")
		m.g.P("Forward: &", stubPackage.Ident("StubForward"), " {")
		m.g.P("ServerAddress: ", strconv.Quote("yourserver:port"), ",")
		m.g.P("Record: true,")
		m.g.P("},")
		m.g.P("},")
	}
	m.g.P("}")
	m.g.P("}")
	m.g.P("")
}

func (m mockServicesGenerator) genGetRequestInstance(service *protogen.Service) {
	m.g.P("func (mock *", unexport(m.getMockServiceName(service)), ") GetRequestInstance(methodName string) ", protoPackage.Ident("Message"), " {")
	m.g.P("switch methodName {")
	for _, method := range service.Methods {
		m.g.P("case ", m.getFullMethodName(service, method), ":")
		m.g.P("return new(", method.Input.GoIdent, ")")
	}
	m.g.P("}")
	m.g.P("return nil")
	m.g.P("}")
	m.g.P("")
}

func (m mockServicesGenerator) genGetResponseInstance(service *protogen.Service) {
	m.g.P("func (mock *", unexport(m.getMockServiceName(service)), ") GetResponseInstance(methodName string) ", protoPackage.Ident("Message"), "{")
	m.g.P("switch methodName {")
	for _, method := range service.Methods {
		m.g.P("case ", m.getFullMethodName(service, method), ":")
		m.g.P("return new(", method.Output.GoIdent, ")")
	}
	m.g.P("}")
	m.g.P("return nil")
	m.g.P("}")
	m.g.P("")
}
func (m mockServicesGenerator) genMockServiceDescriptor(service *protogen.Service) {
	// Service descriptor.
	m.g.P("var ", m.getMockServiceDescriptorName(service), " = ", grpcPackage.Ident("ServiceDesc"), " {")
	m.g.P("ServiceName: ", strconv.Quote(string(service.Desc.FullName())), ",")
	m.g.P("HandlerType: (*", m.getMockServerBaseInterfaceName(service), ")(nil),")
	m.g.P("Methods: []", grpcPackage.Ident("MethodDesc"), "{")
	for _, method := range service.Methods {
		methodHandlerName := m.getMethodHandlerName(service, method)
		if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
			continue // skip if it is streaming
		}
		m.g.P("{")
		m.g.P("MethodName: ", strconv.Quote(string(method.Desc.Name())), ",")
		m.g.P("Handler: ", methodHandlerName, ",")
		m.g.P("},")
	}
	m.g.P("},")
	m.g.P("Streams: []", grpcPackage.Ident("StreamDesc"), "{")
	for _, method := range service.Methods {
		methodHandlerName := m.getMethodHandlerName(service, method)
		if !method.Desc.IsStreamingClient() && !method.Desc.IsStreamingServer() {
			continue
		}
		m.g.P("{")
		m.g.P("StreamName: ", strconv.Quote(string(method.Desc.Name())), ",")
		m.g.P("Handler: ", methodHandlerName, ",")
		if method.Desc.IsStreamingServer() {
			m.g.P("ServerStreams: true,")
		}
		if method.Desc.IsStreamingClient() {
			m.g.P("ClientStreams: true,")
		}
		m.g.P("},")
	}
	m.g.P("},")
	m.g.P("Metadata: \"", m.file.Desc.Path(), "\",")
	m.g.P("}")
	m.g.P()
}

func (m mockServicesGenerator) genMockMethodHandler(service *protogen.Service, method *protogen.Method, hname string) {
	if !method.Desc.IsStreamingClient() && !method.Desc.IsStreamingServer() {
		m.g.P("func ", hname, "(srv interface{}, ctx ", contextPackage.Ident("Context"), ", dec func(interface{}) error, interceptor ", grpcPackage.Ident("UnaryServerInterceptor"), ") (interface{}, error) {")
		m.g.P("in := new(", method.Input.GoIdent, ")")
		m.g.P("if err := dec(in); err != nil { return nil, err }")
		m.g.P("out := new(", method.Output.GoIdent, ")")
		m.g.P("fullMethod := ", m.getFullMethodName(service, method))
		m.g.P("stubsMatcher := (srv).(*", unexport(m.getMockServiceName(service)), ").StubsMatcher")
		m.g.P("return ", grpcHandlerPackage.Ident("MockHandler"), "(ctx, stubsMatcher, fullMethod, in, out)")
		m.g.P("}")
		m.g.P()
		return
	}
	m.g.P("func ", hname, "(srv interface{}, stream ", grpcPackage.Ident("ServerStream"), ") error {")
	m.g.P("// Mock not implemented for streaming")
	m.g.P("return ", fmtPackage.Ident("Errorf"), "(\"mock not implemented for streaming\")")
	m.g.P("}")
	m.g.P()
}

func (m mockServicesGenerator) genGetStubsValidator(service *protogen.Service) {
	m.g.P("func(mock *", unexport(m.getMockServiceName(service)), ") GetStubsValidator() ", stubPackage.Ident("StubsValidator"), "{")
	m.g.P("return mock")
	m.g.P("}")
	m.g.P("")
}

func (m mockServicesGenerator) genIsValid(service *protogen.Service) {

	m.g.P("func(mock *", unexport(m.getMockServiceName(service)), ") IsValid(s *", stubPackage.Ident("Stub"), ") (isValid bool, errorMessages []string) {")
	m.g.P("switch s.FullMethod {")
	for _, method := range service.Methods {
		m.g.P("case ", m.getFullMethodName(service, method), ":")
		m.g.P("req := new(", method.Input.GoIdent, ")")
		m.g.P("resp := new(", method.Output.GoIdent, ")")
		m.g.P("return ", stubPackage.Ident("IsStubValid"), "(s, req.ProtoReflect().Descriptor(), resp.ProtoReflect().Descriptor())")
	}
	m.g.P("default:")
	m.g.P("return true, nil")
	m.g.P("}")
	m.g.P("}")
	m.g.P("")

}

func (m mockServicesGenerator) genForwardRequest(service *protogen.Service) {

	m.g.P("func(mock *", unexport(m.getMockServiceName(service)), ") ForwardRequest(conn grpc.ClientConnInterface, ctx context.Context, methodName string, req interface{}) (interface{}, error) {")
	m.g.P("client := New", service.Desc.Name(), "Client(conn)")
	m.g.P("switch methodName {")
	for _, method := range service.Methods {
		m.g.P("case ", m.getFullMethodName(service, method), ":")
		m.g.P("return client.", method.GoName, "(ctx, req.(*", method.Input.GoIdent, "))")
	}
	m.g.P("}")
	m.g.P("return nil, nil")
	m.g.P("}")
	m.g.P("")

}

func (m mockServicesGenerator) genRemoteClient(service *protogen.Service) {
	m.genRemoteMockClient(service)
	for _, method := range service.Methods {
		m.genRemoteCalls(service, method)
	}
}

func (m mockServicesGenerator) genRemoteCalls(service *protogen.Service, method *protogen.Method) {
	remoteMockClientName := m.getRemoteMockClientName(service)
	// Added the service name to the generated call name to avoid collision in the resulting generated code
	// in case of multiple services with RPCs with the same name
	callName := service.GoName + "_" + method.GoName + "Call"
	methodFullName := m.getFullMethodName(service, method)
	m.g.P("func (c ", remoteMockClientName, ") On", method.GoName, "(ctx ", contextPackage.Ident("Context"), ", request *", method.Input.GoIdent, ")", callName, "{")
	m.g.P("return ", callName, "{")
	m.g.P("ctx: ctx,")
	m.g.P("req: request,")
	m.g.P("client: c,")
	m.g.P("}")
	m.g.P("}")
	m.g.P("")
	m.g.P("type ", callName, " struct {")
	m.g.P("req *", method.Input.GoIdent)
	m.g.P("ctx ", contextPackage.Ident("Context"))
	m.g.P("client ", remoteMockClientName)
	m.g.P("}")
	m.g.P("")
	m.g.P("func (c ", callName, ") Return(response *", method.Output.GoIdent, ") error {")
	m.g.P("return c.client.remoteMockClient.AddStub(", methodFullName, ", c.ctx, c.req, response, nil)")
	m.g.P("}")
	m.g.P("")
	m.g.P("func (c ", callName, ") Error(code ", codesPackage.Ident("Code"), ", message string) error {")
	m.g.P("return c.client.remoteMockClient.AddStub(", methodFullName, ", c.ctx, c.req, nil, ", statusPackage.Ident("New"), "(code, message))")
	m.g.P("}")

	m.g.P("")
}

func (m mockServicesGenerator) genRemoteMockClientClear(service *protogen.Service) {
	remoteMockClientName := m.getRemoteMockClientName(service)
	m.g.P("func (c ", remoteMockClientName, ") Clear() error {")
	m.g.P("return c.remoteMockClient.DeleteAllStubs()")
	m.g.P("}")
	m.g.P("")
}

func (m mockServicesGenerator) genRemoteMockClient(service *protogen.Service) {
	remoteMockClientName := m.getRemoteMockClientName(service)
	m.g.P("func New" + remoteMockClientName + "(")
	m.g.P("host string,")
	m.g.P("port int,")
	m.g.P(") " + remoteMockClientName + "{")
	m.g.P("client := ", remotePackage.Ident("New"), "(host, port)")
	m.g.P("return " + remoteMockClientName + "{")
	m.g.P("host: host,")
	m.g.P("port: port,")
	m.g.P("remoteMockClient: client,")
	m.g.P("}")
	m.g.P("}")
	m.g.P("")
	m.g.P("type ", remoteMockClientName, " struct {")
	m.g.P("remoteMockClient ", remotePackage.Ident("MockServerClient"))
	m.g.P("host string")
	m.g.P("port int")
	m.g.P("}")
	m.g.P("")
	m.genRemoteMockClientClear(service)
}

func (m mockServicesGenerator) getFullMethodName(service *protogen.Service, method *protogen.Method) string {
	return strconv.Quote(fmt.Sprintf("/%s/%s", service.Desc.FullName(), method.GoName))
}

func (m mockServicesGenerator) getMockServiceName(service *protogen.Service) string {
	return service.GoName + "MockService"
}

func (m mockServicesGenerator) getRemoteMockClientName(service *protogen.Service) string {
	return service.GoName + "RemoteMockClient"
}

func (m mockServicesGenerator) getMockServerBaseInterfaceName(service *protogen.Service) string {
	return service.GoName + "Server"
}

func (m mockServicesGenerator) getMockServiceDescriptorName(service *protogen.Service) string {
	return "_" + service.GoName + "_MockServiceDesc"
}

func (m mockServicesGenerator) getMethodHandlerName(service *protogen.Service, method *protogen.Method) string {
	return fmt.Sprintf("_%s_%s_MockHandler", service.GoName, method.GoName)
}

func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }
