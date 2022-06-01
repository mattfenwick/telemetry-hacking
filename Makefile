.PHONY: test fmt vet

test:
	go test ./pkg/...

install:
	go get -u golang.org/x/lint/golint

lint:
	go fmt ./pkg/... ./cmd/...
	#./hack/verify-gofmt.sh
	# ./hack/verify-golint.sh
	./hack/verify-govet.sh

fmt:
	go fmt ./cmd/... ./pkg/...

vet:
	go vet ./cmd/... ./pkg/...
