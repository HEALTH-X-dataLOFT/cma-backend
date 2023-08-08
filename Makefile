BINARY_NAME := cma-backend

build:
	-mkdir _build
	go build -ldflags="-extldflags=-static" -o _build/$(BINARY_NAME) ./cmd/

docker_build:
	docker build --tag cma-backend/cma-backend:dev .

test:
	go test -v ./...

lint:
	golangci-lint run

vulncheck:
	govulncheck ./...
