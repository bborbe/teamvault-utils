
default: precommit

install:
	go build -o $(GOPATH)/bin/teamvault-config-dir-generator cmd/teamvault-config-dir-generator/*
	go build -o $(GOPATH)/bin/teamvault-config-parser cmd/teamvault-config-parser/*
	go build -o $(GOPATH)/bin/teamvault-password cmd/teamvault-password/*
	go build -o $(GOPATH)/bin/teamvault-url cmd/teamvault-url/*
	go build -o $(GOPATH)/bin/teamvault-username cmd/teamvault-username/*
	go build -o $(GOPATH)/bin/teamvault-file cmd/teamvault-file/*

precommit: ensure format generate test check
	@echo "ready to commit"

ensure:
	go mod verify
	go mod vendor

format:
	find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +
	find . -type f -name '*.go' -not -path './vendor/*' -exec go run -mod=vendor github.com/incu6us/goimports-reviser -project-name github.com/bborbe/teamvault-utils -file-path "{}" \;

generate:
	rm -rf mocks avro
	go generate -mod=vendor ./...

test:
	go test -mod=vendor -p=1 -cover -race $(shell go list -mod=vendor ./... | grep -v /vendor/)

check: lint vet errcheck

lint:
	go run -mod=vendor golang.org/x/lint/golint -min_confidence 1 $(shell go list -mod=vendor ./... | grep -v /vendor/)

vet:
	go vet -mod=vendor $(shell go list -mod=vendor ./... | grep -v /vendor/)

errcheck:
	go run -mod=vendor github.com/kisielk/errcheck -ignore '(Close|Write|Fprint)' $(shell go list -mod=vendor ./... | grep -v /vendor/)

addlicense:
	go run -mod=vendor github.com/google/addlicense -c "Benjamin Borbe" -y 2022 -l bsd ./*.go
