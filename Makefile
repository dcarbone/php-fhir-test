tar:
	tar -czf resources.tar.gz -C resources/ .

clean:
	rm resources.tar.gz
	rm bin/php-fhir-test

build: clean tar
	go build -o bin/php-fhir-test