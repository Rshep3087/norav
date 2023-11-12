# Default target
norav: build
	DEBUG=true ./norav

# Build the Go project
build:
	go build -o norav .

# Test the Go project
test:
	go test ./...
