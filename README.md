# PHP FHIR Testing Resources

This repository contains test resources and server to be used by the [php-fhir](https://github.com/dcarbone/php-fhir) 
project.  All other use cases are unsupported.

# Quick Start

You will need:

1. Some way of building containers (Docker, Podman, etc.)
2. Make
3. Golang 1.23+

## Building and Running Binary

```shell
$ make build
$ ./bin/php-fhir-test-server
```

The above will start a webserver bound to localhost port 8080.

# Supported APIs

### `GET /` - List FHIR Versions

```json
["DSTU2","R4","R4B","R5","STU3"]
```

### `GET /$VERSION` - List Embedded Resource Types for Version

With `GET /R4` request:

```json
["CarePlan","CareTeam","Claim","Condition" ...]
```

### `GET /$VERSION/$RESOURCE_TYPE` - List All Resources for Type in Version

With `GET /R4/CarePlan` request:



# Limitations

This project is an extremely simple webserver that embeds and serves a few pre-defined FHIR resources.  It does not
accept new resource submissions.

It is not meant to be performant or in any way fit for use outside its intended use as a test server for the
[php-fhir](https://github.com/dcarbone/php-fhir) project.