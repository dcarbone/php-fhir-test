tar:
	tar -czf resources.tar.gz -C resources/ .

clean:
	rm resources.tar.gz
	rm bin/php-fhir-test

build: tar
	go build -o bin/php-fhir-test

docker-local:
	docker buildx build \
		--load \
		-t dancarbone/php-fhir-test \
		-f docker/Dockerfile \
		.