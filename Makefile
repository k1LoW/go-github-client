LATEST_GO_GITHUB=v39

default: test

ci: test

test:
	cd $(LATEST_GO_GITHUB)/ && go test -v ./... -coverprofile=coverage.out -covermode=count

lint:
	cd $(LATEST_GO_GITHUB)/ && golangci-lint run --config=../.golangci.yml ./...

update:
	scripts/update.sh $(LATEST_GO_GITHUB)
	$(MAKE) copy

copy:
	scripts/copy.sh $(LATEST_GO_GITHUB) v38
	scripts/copy.sh $(LATEST_GO_GITHUB) v37
	scripts/copy.sh $(LATEST_GO_GITHUB) v36
	scripts/copy.sh $(LATEST_GO_GITHUB) v35
	scripts/copy.sh $(LATEST_GO_GITHUB) v34
	scripts/copy.sh $(LATEST_GO_GITHUB) v33
