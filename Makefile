
deps:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/golang/lint/golint
	go get -u github.com/kisielk/errcheck
	go get -u golang.org/x/tools/cmd/goimports

install:
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install cmd/teamvault-config-dir-generator/*.go
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install cmd/teamvault-config-parser/*.go
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install cmd/teamvault-password/*.go
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install cmd/teamvault-url/*.go
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install cmd/teamvault-username/*.go

test:
	go test -cover -race $(shell go list ./... | grep -v /vendor/)

check: format lint vet errcheck

format:
	@go get golang.org/x/tools/cmd/goimports
	@find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +
	@find . -type f -name '*.go' -not -path './vendor/*' -exec goimports -w "{}" +

vet:
	@go vet $(shell go list ./... | grep -v /vendor/)

lint:
	@go get github.com/golang/lint/golint
	@golint -min_confidence 1 $(shell go list ./... | grep -v /vendor/)

errcheck:
	@go get github.com/kisielk/errcheck
	@errcheck -ignore '(Close|Write|Fprint)' $(shell go list ./... | grep -v /vendor/)

