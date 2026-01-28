POLYTOMIC_DEPLOYMENT_URL ?= https://app.polytomic-local.com
TESTARGS ?= -count=1

default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	POLYTOMIC_DEPLOYMENT_URL=$(POLYTOMIC_DEPLOYMENT_URL) TF_ACC=1 go test ./tests/... $(TESTARGS) -timeout 120m
	POLYTOMIC_DEPLOYMENT_URL=$(POLYTOMIC_DEPLOYMENT_URL) TF_ACC=1 go test ./provider/... $(TESTARGS) -timeout 120m
	POLYTOMIC_DEPLOYMENT_URL=$(POLYTOMIC_DEPLOYMENT_URL) TF_ACC=1 go test ./importer/... $(TESTARGS) -timeout 120m

.PHONY: generate-local
generate-local:
	POLYTOMIC_DEPLOYMENT_URL=$(POLYTOMIC_DEPLOYMENT_URL) go generate ./...

.PHONY: dev
dev:
	@echo "==> Setting up environment..."
	./hack/setup_local.sh
	@echo
	@echo "==> Building provider..."
	go generate
	@echo
	./hack/build.sh
	@echo
	@echo "==> Creating templated terraform project..."
	./hack/template.sh
