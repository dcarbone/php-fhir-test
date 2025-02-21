tar:
	tar -czf resources.tar.gz -C resources/ .

clean:
	rm resources.tar.gz
	rm bin/php-fhir-test

build: tar
	go build -o bin/php-fhir-test-server

docker-local:
	docker buildx build \
		--load \
		-t ghcr.io/dcarbone/php-fhir-test-server \
		-f docker/Dockerfile \
		.