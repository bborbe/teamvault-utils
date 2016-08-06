install:
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install bin/k8s_generate_configs/k8s_generate_configs.go
test:
	GO15VENDOREXPERIMENT=1 go test `glide novendor`
format:
	find . -name "*.go" -exec gofmt -w "{}" \;
	goimports -w=true .
prepare:
	npm install
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/Masterminds/glide
	glide install
update:
	glide up
clean:
	rm -rf var vendor
