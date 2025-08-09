README.md:multilistener/* linter/* middleware/* header.in
	gomarkdoc --header-file header.in --output README.md ./...

lint:
	golangci-lint run