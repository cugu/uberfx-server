.PHONY: fmt
fmt:
	@echo "Formatting..."
	templ fmt .
	go fmt ./...
	gci write -s standard -s default -s "prefix(github.com/cugu/uberfx-server)" .
	@echo "Done."

.PHONY: build_uberspace
build_uberspace:
	@echo "Building uberfx-server for uberspace..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/uberfx-server-linux-amd64 ./cmd/uberfx-server
	@echo "Done."
