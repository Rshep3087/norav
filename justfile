# Default target
norav: build
	DEBUG=true ./norav -c .norav.toml

# Build the Go project
build:
	go build -o norav .

# Test the Go project
test:
	go test ./...
