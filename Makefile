install:
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install bin/k8s_generate_configs/k8s_generate_configs.go
test:
	GO15VENDOREXPERIMENT=1 go test `glide novendor`
check:
	golint ./...
	errcheck -ignore '(Close|Write)' ./...
format:
	find . -name "*.go" -exec gofmt -w "{}" \;
	goimports -w=true .
prepare:
	npm install
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/Masterminds/glide
	go get -u github.com/golang/lint/golint
	go get -u github.com/kisielk/errcheck
	glide install
update:
	glide up
clean:
	rm -rf var vendor
