# protoc-gen-mock - A plugin to generate gRPC mock services from protobuf specifications. 

`protoc-gen-mock` is a plugin to generate gRPC mock services based on protobuf specifications. The generated mocks can be used for:

* Mocking a dependent service for testing purposes
* Stub a dependent service during development (while the real service is not implemented yet)

## Installation

It is required to have `protoc-gen-go` installed. Run the following to install it:

```bash
go install github.com/golang/protobuf/protoc-gen-go
```

Then install `protoc-gen-mock` by running:

```
go install github.com/carvalhorr/protoc-gen-mock
```

## Creating mock services

Suppose you have a proto definition called `greeter.proto` with the content:
 
```
syntax = "proto3";

package carvalhorr.greeter;

option go_package = ".;greeter_service";


service Greeter {
	rpc Hello(Request) returns (Response) {}
}

message Request {
	string name = 1;
}

message Response {
	string greeting = 1;
}
```
 
Run the command below to generate the mock.

```bash
protoc --plugin ./protoc-gen-mock --go_out=plugins=grpc:greeter-service --mock_out=greeter-service greeter.proto
```

The command above will generate the files:
```
greeter-service
   greeter.mock.pb.go
   greeter.pb.go
```
## Starting the mock server

Create a file called `greeter.go` with the content:

```
package main

import (
	"github.com/carvalhorr/protoc-gen-mock/bootstrap"
	"github.com/carvalhorr/protoc-gen-mock/grpchandler"
	"github.com/carvalhorr/protoc-gen-mock/stub"
	greetermock "greeter-service" // Import the generated mock service
)

func main() {
	bootstrap.BootstrapServers("./tmp/", 1068, 10010, MockServicesRegistersCallback)
}

var MockServicesRegistersCallback = func(stubsMatcher stub.StubsMatcher) []grpchandler.MockService {
	return []grpchandler.MockService{
		greetermock.NewGreeterService(stubsMatcher), // register the mock service
	}
}
```
Run the mock service:

```
go build
./greeter
```

You will see the message in the console:

```
INFO[2020-04-26T18:13:35+01:00] Supported methods: /carvalhorr.greeter.Greeter/Hello
INFO[2020-04-26T18:13:35+01:00] REST Server listening on port: 1068          
INFO[2020-04-26T18:13:35+01:00] gRPC Server listening on port: 10010    
```

The mock service is listening in two different ports:

* 1068 - A rest service to add/delete/update/get stubs
* 10010 - The gRPC mocked service

## Starting the mock server with support for advanced error mocking

You may need to include the `-trimpath` parameter to the build command if you are using advanced error mocking. In that case, build the program using:

```
go build -trimpath
./greeter
```

## Stubs 

Now you are ready to create the stubs you want the mock service to be able to respond. You can do it using Postman, curl or any other REST client.

Example:

```
POST 127.0.0.1:1068/stubs

{
    "fullMethod": "/carvalhorr.greeter.Greeter/Hello",
    "request": {
        "match": "exact",
        "content": {
            "name": "John"
        }
    },
    "response": {
        "type": "success",
        "content": {
            "greeting": "Hello, John"
        }
    }
}
```

You can verify the stubs that were created with:

```
GET 127.0.0.1:1068/stubs
```

Please refer to the [stubs management API for more details](https://github.com/carvalhorr/protoc-gen-mock/wiki/Managing-stubs-using-the-REST-endpoint).

## Using the mock server
Use your gRPC client to connect to the mock server. By default, it runs on port `10010` and will respond as if it was the real service using the stubs you previously created.

If you created the stub above, now you can make a request to the gRPC method `/carvalhorr.greeter.Greeter/Hello` with the payload `{"name": "John"}` and get the response `{"greeting": "Hello, John"}`.

# More Info

* [Managing stubs through the REST API](https://github.com/carvalhorr/protoc-gen-mock/wiki/Managing-stubs-using-the-REST-API)
* [Advanced Error Mocking](https://github.com/carvalhorr/protoc-gen-mock/wiki/Advanced-error-mocking)