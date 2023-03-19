.PHONY: test
test:
	go test ./... -coverprofile cover.out

.PHONY: bin
bin: fmt vet
	go build -o bin/kubectl-ai

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

lint:
	golangci-lint run -v ./...
