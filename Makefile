#!make
lint:
	gofmt -w -s .
	golangci-lint run core/*

install:
	go get github.com/favadi/protoc-go-inject-tag@latest
	go install github.com/favadi/protoc-go-inject-tag@latest
	go get ./...
	go mod tidy

protogen:
	for proto in services/**/protobuf/*.proto ; do \
		protoc -I/usr/local/include -I. \
		  -I${GOPATH}/src \
		  -I/Users/triviere/lab/googleapis \
		  --go_out=. \
		$$proto ; \
		protoc -I/usr/local/include -I. \
		  -I${GOPATH}/src \
		  -I/Users/triviere/lab/googleapis \
		  --grpc-gateway_out=logtostderr=true:. \
		$$proto ; \
		protoc -I/usr/local/include -I. \
		  -I${GOPATH}/src \
		  -I/Users/triviere/lab/googleapis \
		  --openapiv2_out=logtostderr=true:. \
		$$proto ; \
	done
	@echo "Processing protoc-inject-tag";
	@protoc-go-inject-tag --input="services/*/protobuf/*.pb.go"

clean:
	rm services/**/protobuf/*.pb.go || true
	rm services/**/protobuf/*.pb.gw.go || true
	rm services/**/protobuf/*.swagger.json || true