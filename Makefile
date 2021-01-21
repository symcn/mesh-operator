VERSION ?= v0.0.8
# Image URL to use all building/pushing image targets
IMG_ADDR ?= symcn.tencentcloudcr.com/symcn/meshach
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"
# This repo's root import path (under GOPATH).
ROOT := github.com/meshach

GO_VERSION := 1.14.2
ARCH     ?= $(shell go env GOARCH)
BUILD_DATE = $(shell date +'%Y-%m-%dT%H:%M:%SZ')
COMMIT    = $(shell git rev-parse --short HEAD)
GOENV    := CGO_ENABLED=0 GOOS=$(shell uname -s | tr A-Z a-z) GOARCH=$(ARCH) GOPROXY=https://goproxy.io,direct
GO       := $(GOENV) go build

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	# cd controllers && ginkgo -r --v --randomizeAllSpecs --randomizeSuites --failOnPending -cover -coverprofile=coverage.txt --trace --race --progress -outputdir=.
	go test -v --coverprofile=coverage.txt github.com/symcn/meshach/controllers/...

# Build manager binary
manager: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o bin/meshach -ldflags "-s -w -X  $(ROOT)/pkg/version.Release=$(VERSION) -X  $(ROOT)/pkg/version.Commit=$(COMMIT)   \
	-X  $(ROOT)/pkg/version.BuildDate=$(BUILD_DATE)" cmd/meshach/main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./cmd/meshach/main.go -n sym-admin ctl -v 4

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/operator && kustomize edit set image controller=${IMG_ADDR}:${VERSION}
	cd config/adapter && kustomize edit set image adapter=${IMG_ADDR}:${VERSION}
	kustomize build config/default | kubectl apply -f -

undeploy: manifests
	cd config/operator && kustomize edit set image controller=${IMG_ADDR}:${VERSION}
	cd config/adapter && kustomize edit set image adapter=${IMG_ADDR}:${VERSION}
	kustomize build config/default | kubectl delete -f -


# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=meshach-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build:
	docker build -t ${IMG_ADDR}:${VERSION} .

# Push the docker image
docker-push:
	docker push ${IMG_ADDR}:${VERSION}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
