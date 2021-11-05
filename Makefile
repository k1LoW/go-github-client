LATEST_GO_GITHUB=v39

default: test

test:
	cd $(LATEST_GO_GITHUB)/ && go test -v ./... -coverprofile=coverage.out -covermode=count

lint:
	cd $(LATEST_GO_GITHUB)/ && golangci-lint run ./...

copy:
	scripts/copy.sh $(LATEST_GO_GITHUB) v38
	scripts/copy.sh $(LATEST_GO_GITHUB) v37
	scripts/copy.sh $(LATEST_GO_GITHUB) v36
	scripts/copy.sh $(LATEST_GO_GITHUB) v35
	scripts/copy.sh $(LATEST_GO_GITHUB) v34
	scripts/copy.sh $(LATEST_GO_GITHUB) v33

prerelease:
	git pull origin main --tag
	go mod tidy
	ghch -w -N ${VER}
	gocredits . > CREDITS
	git add CHANGELOG.md CREDITS go.mod
	git commit -m'Bump up version number'
	git tag ${VER}
