test:
	go test -v -timeout 1s -race -bench=. -run=. -coverprofile=coverage.txt -covermode=atomic
