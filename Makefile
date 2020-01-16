test:
	packr2 clean
	go test -failfast -short -cover ./...
	packr2
	go mod tidy -v

cov:
	packr2 clean
	go test -short -coverprofile cover.out ./...
	go tool cover -html cover.out
	packr2
	go mod tidy -v

install:
	packr2 clean
	packr2
	go install -v .

