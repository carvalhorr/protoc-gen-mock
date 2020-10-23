# Forwarding and Recording gRPC Communication

The generated mock server is capable of forwarding an incoming request to another gRPC server, store the response and 
return it to the caller. The recorded responses can be obtained by the `/recordings` endpoint in the REST server.

## Limitations
- Only insecure requests are supported.
- If the same request is made more than once, only the latest response is recorded.
- Error responses are not recorded at this moment.

## Usage

### Creating a forward
To activate forwarding you must include the `forward` instead of the `response` field when creating the stub. Using the 
provided `Greeter` service as an example, the stub would be created like this:

`POST 127.0.0.1:1068/stubs`

```json
{
    "fullMethod": "/carvalhorr.greeter.Greeter/Hello",
    "request": {
        "match": "exact",
        "content": {
            "name": "John"
        }
    },
    "forward": {
        "serverAddress": "yourserver:port",
        "record": true
    }
}
```

You can control if the communication will be recorded through the `record` field. You must provide the remote server 
address in the field `serverAddress`.

When the mock server receive the request that matches the stub's request, it forwards the call to the external server, 
record the request/response if recording is enabled, and return the response to the original caller.

### Obtaining all recordings

To obtain the recorded stubs, make a `GET` call to the `/recordings` endpoint.

## Testing

A sample server for the `Greeter` protobuf is available in `testing/example/server`. To testing the forwarding follow 
these steps:

- Generate the mock server for the provided `Greeter` and start it 
- Start the server `testing/example/server`
- Create a stub with forwarding enabled (see Usage)

Now requests made to mock server will be forwarded to the real gRPC server and the communication will be recorded.

Make a `GET` request to `recordings` to retrieve all the recorded communication.

