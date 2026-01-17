.PHONY: run test testv fmt vet

run:
	go run ./cmd/server

test:
	go test -count=1 ./...

testv:
	go test -count=1 ./... -v

fmt:
	go fmt ./...
	
vet:
	go vet ./...