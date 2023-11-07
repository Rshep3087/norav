homie: build
	./homie

# Build the Go project
build:
	go build -o homie .

# Run the Go project
run:
	go run .

# Test the Go project
test:
	go test ./...
