
default: precommit

install:
	GOBIN=$(GOPATH)/bin go install cmd/teamvault-config-dir-generator/*.go
	GOBIN=$(GOPATH)/bin go install cmd/teamvault-config-parser/*.go
	GOBIN=$(GOPATH)/bin go install cmd/teamvault-password/*.go
	GOBIN=$(GOPATH)/bin go install cmd/teamvault-url/*.go
	GOBIN=$(GOPATH)/bin go install cmd/teamvault-username/*.go
	GOBIN=$(GOPATH)/bin go install cmd/teamvault-file/*.go

precommit: ensure format generate test check
	@echo "ready to commit"

ensure:
	go mod verify
	go mod vendor

format:
	find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +
	find . -type f -name '*.go' -not -path './vendor/*' -exec go run -mod=vendor github.com/incu6us/goimports-reviser -project-name github.com/bborbe/run -file-path "{}" \;

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
	go run -mod=vendor github.com/google/addlicense -c "Benjamin Borbe" -y 2021 -l bsd ./*.go
