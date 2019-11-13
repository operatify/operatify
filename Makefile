# Image URL to use all building/pushing image targets


# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

all: manager

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./main.go

# Install CRDs into a cluster
install: generate
	kubectl apply -f config/crd/bases

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crd/bases
	kustomize build config/default | kubectl apply -f -

timestamp := $(shell /bin/date "+%Y%m%d-%H%M%S")

delete:
	kubectl delete -f config/crd/bases
	kustomize build config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: manifests
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./api/...

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.0
CONTROLLER_GEN=$(shell go env GOPATH)/bin/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

.PHONY: install-bindata
install-bindata:
	go get -u github.com/jteeuwen/go-bindata/...

.PHONE:
generate-template:
	go-bindata -pkg template -prefix pkg/template/assets/ -o pkg/template/templates.go pkg/template/assets/

create-kindcluster:
ifeq (,$(shell kind get clusters))
	@echo "no kind cluster"
else
	@echo "kind cluster is running, deleteing the current cluster"
	kind delete cluster 
endif
	@echo "creating kind cluster"
	kind create cluster

set-kindcluster:
ifeq (${shell kind get kubeconfig-path --name="kind"},${KUBECONFIG})
	@echo "kubeconfig-path points to kind path"
else
	@echo "please run below command in your shell and then re-run make set-kindcluster"
	@echo  "\e[31mexport KUBECONFIG=$(shell kind get kubeconfig-path --name="kind")\e[0m"
	@exit 111
endif
	make create-kindcluster
	
	@echo "getting value of KUBECONFIG"
	@echo ${KUBECONFIG}
	@echo "getting value of kind kubeconfig-path"
	kubectl cluster-info
