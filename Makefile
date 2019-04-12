.SILENT: main development-environment legal list-packages build fast-test test cleanup-functional-tests-environment functional-tests build-integration-tests tag release release-installer promote release-notes-page check-go check-cli-release-config-path
.PHONY: development-environment legal list-packages build fast-test test cleanup-functional-tests-environment functional-tests build-integration-tests tag release release-installer promote release-notes-page
main:
	echo "Liferay CLI build tool commands:"
	echo "development-environment, list-packages, build, fast-test, test, tag, release, promote"
development-environment:
	./scripts/development-environment.sh
legal:
	./scripts/legal.sh
list-packages:
	go list ./...
build:
	go build
fast-test:
	./scripts/test.sh --skip-integration-tests
test:
	./scripts/test.sh
cleanup-functional-tests-environment:
	./scripts/cleanup-functional-tests-environment.sh
functional-tests:
	./scripts/functional-tests.sh
build-integration-tests:
	./scripts/build-integration-tests.sh
tag:
	./scripts/tag.sh
release: check-cli-release-config-path
	./scripts/release.sh --config $$WEDEPLOY_CLI_RELEASE_CONFIG_PATH
release-installer:
	./scripts/release-installer.sh --config $$WEDEPLOY_CLI_RELEASE_CONFIG_PATH
promote: check-cli-release-config-path
	./scripts/promote.sh --config $$WEDEPLOY_CLI_RELEASE_CONFIG_PATH
release-notes-page:
	go run update/releasenotes/page/page.go
check-go:
	if ! which go &> /dev/null; \
	then >&2 echo "Missing dependency: Go is required https://golang.org/"; \
	fi;
check-cli-release-config-path:
	if test -z "$$WEDEPLOY_CLI_RELEASE_CONFIG_PATH"; \
	then >&2 echo "WEDEPLOY_CLI_RELEASE_CONFIG_PATH environment variable is not set"; \
	exit 1; \
	fi;
