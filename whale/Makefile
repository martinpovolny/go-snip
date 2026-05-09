build:
	go build

go_test:
	go test ./...

test:
	./db-cmd.sh delete; ./test_with_curl.sh

docker-build:
	docker build -t whale .

docker-run: docker-build
	docker run -d -p 8081:8080 whale --user 123456

docker-test: docker-run
	./test_with_curl.sh http://localhost:8081

dev-run:
	go run main.go

dev-test:
	./db-cmd.sh delete; ./test_with_curl.sh
