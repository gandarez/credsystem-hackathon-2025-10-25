.PHONY: build-linux-amd64 build clean run test test-docker

build:
	go build -o build/ura-ai .

run:
	go run .

clean:
	rm -rf build/

docker-build:
	docker build -t esquadrao-scheduler:latest .

docker-run:
	docker-compose up

test:
	@echo "Testando API local (porta 8080)..."
	@./test.sh

test-docker:
	@echo "Testando API Docker (porta 18020)..."
	@API_URL=http://localhost:18020 ./test.sh
