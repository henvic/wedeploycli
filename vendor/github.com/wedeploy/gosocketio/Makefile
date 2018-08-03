.SILENT: test get-dev-dependencies
.PHONY: test get-dev-dependencies
test:
	echo "Checking for unchecked errors."
	errcheck $(go list ./...)

	echo "Linting code."
	test -z "$(golint ./... | grep -v "^vendor" | tee /dev/stderr)"

	echo "Examining source code against code defect."
	go vet $(go list ./...)

	echo "Checking if code can be simplified or can be improved."
	megacheck ./...

	echo "Running tests (may take a while)."
	go test $(go list ./...) -race

get-dev-dependencies:
	echo "Installing developer tools."

	echo "cover"
	go get -u golang.org/x/tools/cmd/cover

	echo "goveralls"
	go get -u github.com/mattn/goveralls

	echo "errcheck"
	go get -u github.com/kisielk/errcheck

	echo "golint"
	go get -u github.com/golang/lint/golint

	echo "megacheck"
	go get -u honnef.co/go/tools/cmd/megacheck

	echo "gorilla/websocket"
	go get -u github.com/gorilla/websocket
