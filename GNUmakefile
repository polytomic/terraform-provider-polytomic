POLYTOMIC_DEPLOYMENT_URL ?= app.polytomic-local.com:8443
POLYTOMIC_DEPLOYMENT_KEY ?= secret-key

default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	POLYTOMIC_DEPLOYMENT_URL=$(POLYTOMIC_DEPLOYMENT_URL) POLYTOMIC_DEPLOYMENT_KEY=$(POLYTOMIC_DEPLOYMENT_KEY) TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
