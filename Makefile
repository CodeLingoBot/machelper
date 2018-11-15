build:
	go build -o machelper 
dev: 
	go run main.go
test:
	go test -v -race -cover  ./...