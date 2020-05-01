package main

import (
	"github.com/carvalhorr/protoc-gen-mock/bootstrap"
	"github.com/carvalhorr/protoc-gen-mock/grpchandler"
	"github.com/carvalhorr/protoc-gen-mock/stub"
	testservice "github.com/carvalhorr/protoc-gen-mock/test-service"
)

func main() {
	/*
		spec := &stub.ErrorDetailsSpec{
			Import: "google.golang.org/genproto/googleapis/rpc/errdetails",
			Type:   "PreconditionFailure",
		}
		engine, err := stub.NewCustomErrorEngine("./plugins/")
		if err != nil {
			panic(err)
		}
		instance, err := engine.GetNewInstance(spec)
		if err != nil {
			panic(err)
		}
		fmt.Println(reflect.TypeOf(instance))
		fmt.Println(stub.CreateStubExample(instance.(proto.Message)))
		instance2, err2 := engine.GetNewInstance(spec)
		if err2 != nil {
			panic(err2)
		}
		fmt.Println(reflect.TypeOf(instance2))
		fmt.Println(stub.CreateStubExample(instance2.(proto.Message)))


		fmt.Println(errorextension.CreateHash(spec))
		err = errorextension.GeneratePluginCode(spec)
		fmt.Println(err)
		err = errorextension.CompilePlugin(nil)
		fmt.Println(err)
		fmt.Println(errorextension.LoadType(nil))

	*/
	bootstrap.BootstrapServers("./tmp", 1068, 10010, MockServicesRegistersCallback)
}

var MockServicesRegistersCallback = func(stubsMatcher stub.StubsMatcher) []grpchandler.MockService {
	return []grpchandler.MockService{
		testservice.NewTestProtobufMockService(stubsMatcher),
	}
}
