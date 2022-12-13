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
ZIPKIN       := openzipkin/zipkin:2.23
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
		-t sales-api:$(VERSION) \
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

dev-up:
	kind create cluster \
		--image kindest/node:v1.26.0@sha256:${KIND_NODE_DIGEST} \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/dev/kind-config.yaml	

dev-load:
	cd zarf/k8s/kind/sales; kustomize edit set image sales-api-image=sales-api-${GOARCH}:${VERSION}
	# kind load docker-image sales-${GOARCH}:${VERSION} --name ${KIND_CLUSTER}