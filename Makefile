.PHONY: all test lint clean bench coverage proto-gen proto-clean integration

all: test lint

test:
	go test -race -v ./...

integration:
	go test -tags=integration -v ./integration/

lint:
	golangci-lint run

clean:
	go clean
	rm -rf coverage.out

bench:
	go test -bench=. -benchmem ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out 

proto-gen:
	protoc --go_out=proto/generated --go-grpc_out=proto/generated --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/service.proto

proto-clean:
	rm -rf proto/generated/* 