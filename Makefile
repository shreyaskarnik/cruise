PROJECT = cruise
REGISTRY ?= gcr.io/heptio-images
IMAGE := $(REGISTRY)/$(PROJECT)

GIT_REF = $(shell git rev-parse --short=8 --verify HEAD)
VERSION ?= $(GIT_REF)
PKGS := $(shell go list ./...)

test: install
	go test ./... -cover

check: test vet staticcheck unused
	@echo Checking code is gofmted
	@bash -c 'if [ -n "$(gofmt -s -l .)" ]; then echo "Go code is not formatted:"; gofmt -s -d -e .; exit 1;fi'

install:
	go install -v -tags "oidc gcp" ./...

container:
	docker build . -t $(IMAGE):$(VERSION)

push: container
	docker push $(IMAGE):$(VERSION)
	@if git describe --tags --exact-match >/dev/null 2>&1; \
	then \
	    docker tag $(IMAGE):$(VERSION) $(IMAGE):latest; \
	    docker push $(IMAGE):latest; \
	fi

vet: test
	go vet ./...

staticcheck:
	@go get honnef.co/go/tools/cmd/staticcheck
	staticcheck $(PKGS)

unused:
	@go get honnef.co/go/tools/cmd/unused
	unused $(PKGS)
