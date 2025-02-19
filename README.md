# PHP FHIR Testing Resources

This repository contains test resources and server to be used by the [php-fhir](https://github.com/dcarbone/php-fhir)
project. All other use cases are unsupported.

<!-- TOC -->
* [PHP FHIR Testing Resources](#php-fhir-testing-resources)
* [Generated Resources](#generated-resources)
    * [Data Sources](#data-sources)
  * [Building and Running Binary](#building-and-running-binary)
  * [Building and Running Local Docker](#building-and-running-local-docker)
* [Supported APIs](#supported-apis)
    * [Note on XML Format](#note-on-xml-format)
<!-- TOC -->

# Generated Resources

This little webserver embeds any / all resources contained under the [./resources](./resources) directory. These are
static and this webserver is (currently) read-only.

### Data Sources

Resources were generated / sourced from the below sources:

* [synthetichealth/synthea](https://github.com/synthetichealth/synthea) project, using the seed `9001`.
* Resources found on the open [HAPI](https://hapifhir.io/) servers.
* Other servers listed on the [FHIR Public Test Servers](https://confluence.hl7.org/display/FHIR/Public+Test+Servers) page. 

## Building and Running Binary

```shell
$ make build
$ ./bin/php-fhir-test-server
```

The above command compiles and runs the `php-fhir-test-server` binary. The webserver binds to `127.0.0.1:8080` by
default, but you may change this by providing the `-bind` flag at runtime.

## Building and Running Local Docker

```shell
$ make docker-local
$ docker run --rm -p '127.0.0.1:8080:8080' dancarbone/php-fhir-test-server
```

The above command compiles and runs a new container image, binding `127.0.0.1:8080` on your host.  You can change
the port bound on your host by changing the final `:8080` segment in the value passed to `-p`.

# Supported APIs

| Route                                       | Parameters                             | Description                                                           |
|---------------------------------------------|----------------------------------------|-----------------------------------------------------------------------|
| `GET /`                                     | -                                      | List available FHIR versions.                                         |
| `GET /$VERSION`                             | -                                      | List available resources for FHIR version.                            |
| `GET /$VERSION/$RESOURCE_TYPE`              | `_count=[0,...]`, `_format=[xml,json]` | Retrieve`Bundle` with one or more resources for a particular version. |
| `GET /$VERSION/$RESOURCE_TYPE/$RESOURCE_ID` | `_format=[xml,json]`                   | Retrieve a specific resource.                                         |

### Note on XML Format

The XML formatting provided by this server is a work in progress, and may not be 100% compliant.

Opening an issue with examples of errs is welcome!

