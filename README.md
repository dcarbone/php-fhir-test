# PHP FHIR Testing Resources

This repository contains test resources and server to be used by the [php-fhir](https://github.com/dcarbone/php-fhir) 
project.  All other use cases are unsupported.

# Generated Resources

This little webserver embeds any / all resources contained under the [./resources](./resources) directory.  These are
static and this webserver is (currently) read-only.

### Data Sources

Resources were generated / sourced from the below sources:

* [synthetichealth/synthea](https://github.com/synthetichealth/synthea) project, using the seed `9001`.

## Building and Running Binary

```shell
$ make build
$ ./bin/php-fhir-test-server
```

The above command compiles and runs the `php-fhir-test-server` binary with defaults.  The webserver binds
to `127.0.0.1:8080`.

## Building and Running Local Docker

```shell
$ make docker-local
$ docker run --rm -p '127.0.0.1:8080:8080' dancarbone/php-fhir-test-server
```

The above command compiles and runs a new container image, binding `127.0.0.1:8080` on your host.

# Supported APIs

| Route                                       | Parameters | Description                                                           |
|---------------------------------------------|------------|-----------------------------------------------------------------------|
| `GET /`                                     | -          | List available FHIR versions.                                         |
| `GET /$VERSION`                             | -          | List available resources for FHIR version.                            |
| `GET /$VERSION/$RESOURCE_TYPE`              | `_count`   | Retrieve`Bundle` with one or more resources for a particular version. |
| `GET /$VERSION/$RESOURCE_TYPE/$RESOURCE_ID` | -          | Retrieve a sepcific resource.                                         |
