.PHONY: run test testv race fmt vet

run:
	go run ./cmd/server

test:
	go test -count=1 ./...

testv:
	go test -count=1 ./... -v

race:
	go test -count=1 -race ./...

fmt:
	go fmt ./...
	
vet:
	go vet ./...
