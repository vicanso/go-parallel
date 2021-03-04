lint:
	golangci-lint run

test:
	go test -race -cover ./...

test-cover:
	go test -race -coverprofile=test.out ./... && go tool cover --html=test.out