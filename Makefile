.PHONY: all
all: lint test coverage check-licenses build

.PHONY: build
build:

.PHONY: lint
lint: prepare
	golangci-lint run

.PHONY: prepare
prepare:
	go mod tidy

cmd1 cmd2:
	go build -ldflags "-X github.com/Eyevinn/{Name}/internal.commitVersion=$$(git describe --tags HEAD) -X github.com/Eyevinn/{Name}/internal.commitDate=$$(git log -1 --format=%ct)" -o out/$@ ./cmd/$@/main.go

.PHONY: test
test: prepare
	go test ./...

.PHONY: coverage
coverage:
	go test ./... -coverprofile coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func coverage.out -o coverage.txt
	tail -1 coverage.txt

.PHONY: clean
clean:
	rm -f out/*
	rm -r examples-out/*

.PHONY: install
install: all
	cp out/* $(GOPATH)/bin/

.PHONY: update
update:
	go get -t -u ./...

.PHONY: check-licenses
check-licenses: prepare
	wwhrd check
