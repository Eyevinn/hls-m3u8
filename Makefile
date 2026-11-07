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

# wwhrd reads the licenses of the packages in a vendor directory, and without one it checks nothing and passes.
# So the Go files, go.mod, go.sum and .wwhrd.yml are copied to a temporary directory, the dependencies are
# vendored there, and wwhrd checks that copy, which leaves no vendor directory in the repository.
.PHONY: check-licenses
check-licenses: prepare
	@tmp=$$(mktemp -d) && trap 'rm -rf "$$tmp"' EXIT && \
	git ls-files -z -c -o --exclude-standard '*.go' go.mod go.sum .wwhrd.yml | xargs -0 tar cf - | tar xf - -C "$$tmp" && \
	cd "$$tmp" && go mod vendor && wwhrd -q check
