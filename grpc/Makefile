GOPATH = /Users/martin.povolny/go
# $(go env GOPATH)

.PHONY: proto
proto:
	protoc -I. -I$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis --go_out=. --go-grpc_out=. --grpc-gateway_out=. example.proto

server:
	go run server.go

client:
	go run client.go

grpcurl:
	grpcurl -plaintext localhost:50051 list
	grpcurl -plaintext localhost:50051 describe example.ExampleService
	grpcurl -plaintext -d '{"name": "grpcurl"}' localhost:50051 example.ExampleService/SayHello

