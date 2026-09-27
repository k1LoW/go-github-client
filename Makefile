BASE_GO_GITHUB = 33
LATEST_GO_GITHUB ?= 90

default: test

ci: test

test:
	cd v$(BASE_GO_GITHUB)/ && go test -v ./... -coverprofile=coverage.out -covermode=count

lint:
	cd v$(BASE_GO_GITHUB)/ && golangci-lint run --config=../.golangci.yml ./...

# v87 is skipped because go-github v87 only offers WithEnterpriseURLs, which forces an /api/v3/ suffix on arbitrary endpoints.
update:
	for i in {34..86}; do scripts/copy.sh v$(BASE_GO_GITHUB) v$$i; done
	scripts/copy.sh v86 v$(BASE_GO_GITHUB)
	for i in {89..$(LATEST_GO_GITHUB)}; do scripts/copy.sh v88 v$$i; done
	scripts/copy.sh v$(LATEST_GO_GITHUB) v88

release:
	git push origin main --tag
