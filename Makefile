BASE_GO_GITHUB=33
LATEST_GO_GITHUB=45

default: test

ci: test

test:
	cd v$(BASE_GO_GITHUB)/ && go test -v ./... -coverprofile=coverage.out -covermode=count

lint:
	cd v$(BASE_GO_GITHUB)/ && golangci-lint run --config=../.golangci.yml ./...

update:
	for i in {34..$(LATEST_GO_GITHUB)}; do scripts/copy.sh v$(BASE_GO_GITHUB) v$$i; done
	scripts/copy.sh v$(LATEST_GO_GITHUB) v$(BASE_GO_GITHUB)
