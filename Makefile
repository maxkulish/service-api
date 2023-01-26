SHELL := /bin/bash

export GO111MODULE=on
export CGO_ENABLED=0
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
GOPATH=$(shell go env GOPATH)
ifeq (,$(shell go env GOBIN))
GOBIN=$(GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# ==============================================================================
# Install dependencies

GOLANG       := golang:1.19
ALPINE       := alpine:3.17
KIND         := kindest/node:v1.25.3
POSTGRES     := postgres:15-alpine
VAULT        := hashicorp/vault:1.12
ZIPKIN       := openzipkin/zipkin:latest
TELEPRESENCE := docker.io/datawire/tel2:2.9.3

dev.setup.mac.common:
	brew update
	brew tap hashicorp/tap
	brew list kind || brew install kind
	brew list kubectl || brew install kubectl
	brew list kustomize || brew install kustomize
	brew list pgcli || brew install pgcli
	brew list vault || brew install vault

dev.setup.mac: dev.setup.mac.common
	brew datawire/blackbird/telepresence || brew install datawire/blackbird/telepresence

dev.setup.mac.arm64: dev.setup.mac.common
	brew datawire/blackbird/telepresence-arm64 || brew install datawire/blackbird/telepresence-arm64

dev.docker:
	docker pull $(GOLANG)
	docker pull $(ALPINE)
	docker pull $(KIND)
	docker pull $(POSTGRES)
	docker pull $(VAULT)
	docker pull $(ZIPKIN)
	docker pull $(TELEPRESENCE)

# ==============================================================================
# Building containers

# $(shell git rev-parse --short HEAD)
VERSION := 1.5

all: sales

sales:
	docker build \
		-f zarf/docker/dockerfile.sales-api \
		-t sales-api-${GOARCH}:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ==============================================================================
# Running from within k8s/kind

KIND_CLUSTER := sales-cluster

ifeq ($(GOARCH), arm64)
KIND_NODE_DIGEST ?= 608b47ae5233bb3ff28f9ce5fea24f869639718dac8b26855aba13187bf690a4
else
KIND_NODE_DIGEST ?= 7998effe843cbcb88bc6876a142437e7bccf6d77c5a928dd2325f2ff6fee6f60
endif

.PHONY: dev-up dev-down dev-load dev-apply dev-status dev-status-sales dev-status-db dev-restart dev-update dev-update-apply dev-logs monitor load
dev-up:
	kind create cluster \
		--image kindest/node:v1.26.0@sha256:${KIND_NODE_DIGEST} \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/dev/kind-config.yaml
		
	kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER)
	kubectl config set-context --current --namespace=sales-system

dev-down:
	kind delete cluster --name $(KIND_CLUSTER)

dev-load:
	cd zarf/k8s/dev/sales; kustomize edit set image sales-api-image=sales-api-${GOARCH}:${VERSION}
	kind load docker-image sales-api-${GOARCH}:${VERSION} --name ${KIND_CLUSTER}

dev-apply:
	kustomize build zarf/k8s/dev/database-pod | kubectl apply -f -
	kubectl wait --namespace=database-system --timeout=120s --for=condition=Available deployment/database-pod

	kustomize build zarf/k8s/dev/zipkin | kubectl apply -f -
	kubectl wait --timeout=120s --namespace=sales-system --for=condition=Available deployment/zipkin

	kustomize build zarf/k8s/dev/sales | kubectl apply -f -
	kubectl wait --timeout=120s --namespace=sales-system --for=condition=Available deployment/sales

dev-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

dev-status-sales:
	kubectl get pods -o wide --watch

dev-status-db:
	kubectl get pods -o wide --watch --namespace=database-system

dev-restart:
	kubectl rollout restart deployment sales --namespace=sales-system

dev-update: all dev-load dev-restart 

dev-update-apply: all dev-load dev-apply

dev-logs:
	kubectl logs --namespace=sales-system -l app=sales --all-containers=true -f --tail=100 --max-log-requests=6 | go run app/tooling/logfmt/main.go -service=SALES-API

dev-logs-zipkin:
	kubectl logs --namespace=sales-system -l app=zipkin --all-containers=true -f --tail=100

dev-services-delete:
	kustomize build zarf/k8s/dev/sales | kubectl delete -f -
	kustomize build zarf/k8s/dev/zipkin | kubectl delete -f -
	kustomize build zarf/k8s/dev/database | kubectl delete -f -

# ==============================================================================
# Testing running system
# Install expvarmon before
# go install github.com/divan/expvarmon
monitor:
	expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

# Install hey before
# brew install hey
load:
	hey -m GET -c 100 -n 1000 http://localhost:3000/v1/test

# ==============================================================================
# Modules support

.PHONY: deps deps-reset tidy deps-list deps-upgrade deps-cleancache list run admin gen-keys-openssl test test-verbose database
deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

deps-list:
	go list -m -u -mod=readonly all

deps-upgrade:
	go get -u -v ./...
	go mod tidy
	go mod vendor

deps-cleancache:
	go clean -modcache

list:
	go list -mod=mod all

run:
	go run app/services/sales-api/main.go | go run app/tooling/logfmt/main.go

admin:
	go run app/tooling/admin/main.go

gen-keys-openssl:
	openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:4096
	openssl rsa -pubout -in private.pem -out public.pem

# -count=1 is needed to ignore cached test results
test:
	go test ./... -count=1 
	staticcheck -checks=all ./...

test-verbose:
	go test -v ./... -count=1 -coverprofile=coverage.out -covermode=atomic
	staticcheck -checks=all ./...

database:
	dblab --host 0.0.0.0 --user postgres --db postgres --pass postgres --ssl disable --port 5432 --driver postgres

# Testing Auth
# curl -il http://localhost:3000/v1/test-auth
# curl -il -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/test-auth

# curl -il --user "admin@example.com:gophers" http://sales-service.sales-system.svc.cluster.local:3000/v1/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
# export TOKEN="COPY TOKEN STRING FROM LAST CALL"
# curl -il -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/test-auth
#
# For testing load on the service.
# hey -m GET -c 100 -n 10000 -H "Authorization: Bearer ${TOKEN}" http://sales-service.sales-system.svc.cluster.local:3000/v1/users/1/2
#