POLYTOMIC_DEPLOYMENT_URL ?= https://app.polytomic-local.com:8443

default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	POLYTOMIC_DEPLOYMENT_URL=$(POLYTOMIC_DEPLOYMENT_URL) TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m


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
