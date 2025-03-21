all: test vet staticcheck

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

staticcheck:
	staticcheck ./...