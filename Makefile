.SILENT: main get-dependencies list-packages build fast-test test build-integration-tests build-functional-tests tag release promote check-go check-cli-release-config-path
.PHONY: get-dependencies list-packages build fast-test test build-integration-tests build-functional-tests tag release promote
main:
	echo "WeDeploy CLI build tool commands:"
	echo "get-dependencies, list-packages, build, fast-test, test, build-functional-tests, tag, release, promote"
get-dependencies: check-go
	if ! which dep &> /dev/null; \
	then >&2 echo "Install dep to manage dependencies with go get -u github.com/golang/dep/cmd/dep"; \
	fi;

	dep ensure
list-packages:
	go list ./... | grep -v /vendor/
build:
	go build
fast-test:
	./scripts/tag.sh --dry-run --skip-integration-tests
test:
	./scripts/tag.sh --dry-run
build-integration-tests:
	./scripts/build-integration-tests.sh
build-functional-tests:
	./scripts/build-functional-tests.sh
tag:
	./scripts/tag.sh
release: check-cli-release-config-path
	./scripts/release.sh --config $$WEDEPLOY_CLI_RELEASE_CONFIG_PATH
promote: check-cli-release-config-path
	./scripts/promote.sh --config $$WEDEPLOY_CLI_RELEASE_CONFIG_PATH
check-go:
	if ! which go &> /dev/null; \
	then >&2 echo "Missing dependency: Go is required https://golang.org/"; \
	fi;
check-cli-release-config-path:
	if test -z "$$WEDEPLOY_CLI_RELEASE_CONFIG_PATH"; \
	then >&2 echo "WEDEPLOY_CLI_RELEASE_CONFIG_PATH environment variable is not set"; \
	exit 1; \
	fi;
