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
  * [GET /](#get-)
  * [GET /{fhir_version}](#get-fhir_version)
  * [GET /{fhir_version}/{resource_type}](#get-fhir_versionresource_type)
  * [GET /{fhir_version}/{resource_type}/{resource_id}](#get-fhir_versionresource_typeresource_id)
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
$ docker run --rm -p '127.0.0.1:8080:8080' ghcr.io/dcarbone/php-fhir-test-server
```

The above command compiles and runs a new container image, binding `127.0.0.1:8080` on your host.  You can change
the port bound on your host by changing the final `:8080` segment in the value passed to `-p`.

# Supported APIs

Below is the list of supported API's, along with an example query & response of each.

## GET /

Returns list of all available FHIR versions with resources.

**Parameters:**

| Name      | Type      | Default | Description                                                          |
|-----------|-----------|---------|----------------------------------------------------------------------|
| `_pretty` | `boolean` | `false` | If set and not equal to `false`, "pretty prints" the output          |
| `_format` | `string`  | -       | Must be one of: `[xml json]`.  May also be set with `Accept` header. |

**Example:**

```shell
curl -L -H 'Accept: application/json' http://127.0.0.1:8080/\?_pretty
```

<details>
  <summary>JSON example</summary>

  ```json
[
  "DSTU1",
  "DSTU2",
  "STU3",
  "R4",
  "R4B",
  "R5"
]
  ```
</details>

```shell
curl -L -H 'Accept: application/xml' http://127.0.0.1:8080/\?_pretty
```

<details>
  <summary>XML example</summary>

  ```xml
<?xml version="1.0" encoding="UTF-8"?>
<FHIRVersions>
  <FHIRVersion value="DSTU1"></FHIRVersion>
  <FHIRVersion value="DSTU2"></FHIRVersion>
  <FHIRVersion value="STU3"></FHIRVersion>
  <FHIRVersion value="R4"></FHIRVersion>
  <FHIRVersion value="R4B"></FHIRVersion>
  <FHIRVersion value="R5"></FHIRVersion>
</FHIRVersions>
  ```
</details>

## GET /{fhir_version}

Returns a list of available resource types for a given FHIR version.

**Parameters:**

| Name      | Type      | Default | Description                                                          |
|-----------|-----------|---------|----------------------------------------------------------------------|
| `_pretty` | `boolean` | `false` | If set and not equal to `false`, "pretty prints" the output          |
| `_format` | `string`  | -       | Must be one of: `[xml json]`.  May also be set with `Accept` header. |

**Example:**

```shell
curl -L -H 'Accept: application/json' http://127.0.0.1:8080/R4\?_pretty
```

<details>
  <summary>JSON example</summary>

  ```json
{
  "fhirVersion": 4,
  "resources": [
    "CarePlan",
    "CareTeam",
    "Claim",
    "Condition",
    "Device",
    "DiagnosticReport",
    "DocumentReference",
    "Encounter",
    "ExplanationOfBenefit",
    "ImagingStudy",
    "Immunization",
    "Location",
    "Medication",
    "MedicationAdministration",
    "MedicationRequest",
    "Observation",
    "Organization",
    "Parameters",
    "Patient",
    "Practitioner",
    "PractitionerRole",
    "Procedure",
    "Provenance",
    "SupplyDelivery"
  ]
}

  ```
</details>

```shell
curl -L -H 'Accept: application/xml' http://127.0.0.1:8080/R4\?_pretty
```

<details>
  <summary>XML example</summary>

  ```xml
<?xml version="1.0" encoding="UTF-8"?>
<FHIRVersion value="R4">
  <Resources>
    <Resource value="CarePlan"></Resource>
    <Resource value="CareTeam"></Resource>
    <Resource value="Claim"></Resource>
    <Resource value="Condition"></Resource>
    <Resource value="Device"></Resource>
    <Resource value="DiagnosticReport"></Resource>
    <Resource value="DocumentReference"></Resource>
    <Resource value="Encounter"></Resource>
    <Resource value="ExplanationOfBenefit"></Resource>
    <Resource value="ImagingStudy"></Resource>
    <Resource value="Immunization"></Resource>
    <Resource value="Location"></Resource>
    <Resource value="Medication"></Resource>
    <Resource value="MedicationAdministration"></Resource>
    <Resource value="MedicationRequest"></Resource>
    <Resource value="Observation"></Resource>
    <Resource value="Organization"></Resource>
    <Resource value="Parameters"></Resource>
    <Resource value="Patient"></Resource>
    <Resource value="Practitioner"></Resource>
    <Resource value="PractitionerRole"></Resource>
    <Resource value="Procedure"></Resource>
    <Resource value="Provenance"></Resource>
    <Resource value="SupplyDelivery"></Resource>
  </Resources>
</FHIRVersion>
  ```
</details>


## GET /{fhir_version}/{resource_type}

Returns a `Bundle` of resources of a particular type for a particular FHIR version.

**Parameters:**

| Name      | Type      | Default | Description                                                                         |
|-----------|-----------|---------|-------------------------------------------------------------------------------------|
| `_pretty` | `boolean` | `false` | If set and not equal to `false`, "pretty prints" the output                         |
| `_format` | `string`  | -       | Must be one of: `[xml json]`.  May also be set with `Accept` header.                |
| `_count`  | `integer` | 0       | Maximum number of resources to return.  Value must be >= 0, and 0 means return all. |

**Query Example:**

```shell
curl -L -H 'Accept: application/json' http://127.0.0.1:8080/R4/Patient\?_pretty\&_count\=2
```

```shell
curl -L -H 'Accept: application/xml' http://127.0.0.1:8080/R4/Patient\?_pretty\&_count\=2
```

## GET /{fhir_version}/{resource_type}/{resource_id}

Returns a `Bundle` of resources of a particular type for a particular FHIR version.

**Parameters:**

| Name      | Type      | Default | Description                                                                         |
|-----------|-----------|---------|-------------------------------------------------------------------------------------|
| `_pretty` | `boolean` | `false` | If set and not equal to `false`, "pretty prints" the output                         |
| `_format` | `string`  | -       | Must be one of: `[xml json]`.  May also be set with `Accept` header.                |

**Query Example:**

```shell
curl -L -H 'Accept: application/json' http://127.0.0.1:8080/R4/Patient/2028cf85-bdf0-4a2e-2f82-5f3a37078971\?_pretty
```

```shell
curl -L -H 'Accept: application/xml' http://127.0.0.1:8080/R4/Patient/2028cf85-bdf0-4a2e-2f82-5f3a37078971\?_pretty
```
