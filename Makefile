
all: test install

install:
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install cmd/teamvault-config-dir-generator/*.go
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install cmd/teamvault-config-parser/*.go
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install cmd/teamvault-password/*.go
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install cmd/teamvault-username/*.go
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install cmd/teamvault-url/*.go

glide:
	go get github.com/Masterminds/glide

test: glide
	GO15VENDOREXPERIMENT=1 go test -cover `glide novendor`

vet:
	go tool vet .
	go tool vet --shadow .

lint:
	golint -min_confidence 1 ./...

errcheck:
	errcheck -ignore '(Close|Write)' ./...

check: lint vet errcheck

goimports:
	go get golang.org/x/tools/cmd/goimports

format: goimports
	find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +
	find . -type f -name '*.go' -not -path './vendor/*' -exec goimports -w "{}" +

prepare:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/golang/lint/golint
	go get -u github.com/kisielk/errcheck
	go get -u github.com/Masterminds/glide
