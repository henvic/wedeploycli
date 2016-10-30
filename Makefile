.SILENT: main get-dependencies list-packages build test build-functional-tests release promote check-go check-cli-release-config-path
.PHONY: get-dependencies list-packages build test build-functional-tests release promote
main:
	echo "WeDeploy CLI build tool commands:"
	echo "get-dependencies, list-packages, build, test, build-functional-tests, release, promote"
get-dependencies: check-go
	if ! which glide &> /dev/null; \
	then >&2 echo "Missing dependency: Glide is required https://glide.sh/"; \
	fi;

	glide install
list-packages:
	go list ./... | grep -v /vendor/
build:
	go build
test:
	./scripts/release.sh --pre-release
build-functional-tests:
	./scripts/build-functional-tests.sh
release: check-cli-release-config-path
	./scripts/release.sh $$WEDEPLOY_CLI_RELEASE_CONFIG_PATH
promote: check-cli-release-config-path
	./scripts/promote.sh $$WEDEPLOY_CLI_RELEASE_CONFIG_PATH
check-go:
	if ! which go &> /dev/null; \
	then >&2 echo "Missing dependency: Go is required https://golang.org/"; \
	fi;
check-cli-release-config-path:
	if test -z "$$WEDEPLOY_CLI_RELEASE_CONFIG_PATH"; \
	then >&2 echo "WEDEPLOY_CLI_RELEASE_CONFIG_PATH environment variable is not set"; \
	exit 1; \
	fi;
