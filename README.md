#protoc-gen-mock

## Install protoc-gen-go

It is required to have `protoc-gen-go` installed.

```bash
go install github.com/golang/protobuf/protoc-gen-go
```

## Usage
Clone the repo:

```
git clone https://github.com/carvalhorr/protoc-gen-mock.git
```

Build:

```bash
go build
```

Execute:


```bash
protoc --plugin ./protoc-gen-mock --go_out=plugins=grpc:test-service --mock_out=test-service test.proto
```

The command above will generate the files:
```
test-service
   test.mock.pb.go
   test.pb.go
```
## Run the example

From  the `example` folder, run:

```
go run example.go
```

You will see the message in the console:

```
INFO[2020-04-26T18:13:35+01:00] Supported methods: /carvalhorr.proto.test.TestProtobuf/GetProtoTest 
INFO[2020-04-26T18:13:35+01:00] REST Server listening on port: 1068          
INFO[2020-04-26T18:13:35+01:00] gRPC Server listening on port: 10010    
```

## Create stubs 

Use Postman or curl to create stubs. Example:
```
POST 127.0.0.1:1068/stubs

{
    "fullMethod": "/carvalhorr.proto.test.TestProtobuf/GetProtoTest",
    "request": {
        "match": "exact",
        "content": {
            "name": "Doe"
        },
        "metadata": {}
    },
    "response": {
        "type": "success",
        "content": {
            "name": "John Doe"
        }
    }
}
```

You can verify the stubs that were created with:

```
GET 127.0.0.1:1068/stubs
```

If you created the example above, you will see this response:
```
[
    {
        "fullMethod": "/carvalhorr.proto.test.TestProtobuf/GetProtoTest",
        "request": {
            "match": "exact",
            "content": {
                "name": "Doe"
            },
            "metadata": {}
        },
        "response": {
            "type": "success",
            "content": {
                "name": "John Doe"
            },
            "error": ""
        }
    }
]
```
## Using the mock server
Use any gRPC client to connect to the mock server, By default ir runs on port `10010`. It will respond as if it was the real service using the stubs you first loaded.

You can register more than one mock service to the same server. See the example below:

```
package main

import (
	"github.com/carvalhorr/protoc-gen-mock/bootstrap"
	"github.com/carvalhorr/protoc-gen-mock/grpchandler"
	"github.com/carvalhorr/protoc-gen-mock/stub"
	testservice "github.com/carvalhorr/protoc-gen-mock/test-service"
    // Import your generated mock server here
)

func main() {
	bootstrap.BootstrapServers(1068, 10010, MockServicesRegistersCallback)
}

var MockServicesRegistersCallback = func(stubsMatcher stub.StubsMatcher) []grpchandler.MockService {
	return []grpchandler.MockService{
		testservice.NewTestProtobufMockService(stubsMatcher),
        // call the function to create the generated mock service
	}
}
```


## Use metadata for matching
You can also use metadata in the request to select which stub will be selected to provide the response.
