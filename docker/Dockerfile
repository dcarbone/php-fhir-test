FROM golang:1.24-alpine AS build

LABEL org.opencontainers.image.authors="Daniel Carbone<daniel.p.carbone@gmail.com>"
LABEL org.opencontainers.image.url="https://github.com/dcarbone/php-fhir-test"
LABEL org.opencontainers.image.source="https://github.com/dcarbone/php-fhir-test"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.title="PHP FHIR Test Server Build Image"
LABEL org.opencontainers.image.description="PHP FHIR test API server"

RUN apk add --update make

WORKDIR /app

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make build

RUN ls -l bin

FROM alpine:3

LABEL org.opencontainers.image.authors="Daniel Carbone<daniel.p.carbone@gmail.com>"
LABEL org.opencontainers.image.url="https://github.com/dcarbone/php-fhir-test"
LABEL org.opencontainers.image.source="https://github.com/dcarbone/php-fhir-test"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.title="PHP FHIR Test Server Image"
LABEL org.opencontainers.image.description="PHP FHIR test API server"

COPY --from=build /app/bin/php-fhir-test-server /php-fhir-test-server

HEALTHCHECK --interval=1s --timeout=1s --start-period=5s \
    CMD nc -z 127.0.0.1:8080

ENTRYPOINT [ "/php-fhir-test-server", "-bind", "0.0.0.0:8080" ]