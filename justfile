# Default target
homie: build
	DEBUG=true ./homie

# Build the Go project
build:
	go build -o homie .

# Test the Go project
test:
	go test ./...
