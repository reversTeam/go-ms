#!make
lint:
	gofmt -w -s .
	golangci-lint run examples/*
	golangci-lint run services/*
	golangci-lint run core/*
	golangci-lint run helpers/*

install:
	go get ./...

protogen:
	for proto in services/**/protobuf/*.proto ; do \
		protoc -I/usr/local/include -I. \
		  -I${GOPATH}/src \
		  -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		  --go_out=plugins=grpc:. \
		$$proto ; \
		protoc -I/usr/local/include -I. \
		  -I${GOPATH}/src \
		  -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		  --grpc-gateway_out=logtostderr=true:. \
		$$proto ; \
		protoc -I/usr/local/include -I. \
		  -I${GOPATH}/src \
		  -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		  --swagger_out=logtostderr=true:. \
		$$proto ; \
	done

clean:
	rm services/**/protobuf/*.pb.go || true
	rm services/**/protobuf/*.pb.gw.go || true
	rm services/**/protobuf/*.swagger.json || true