FROM golang:1.23-alpine AS build

RUN apk add --update make

WORKDIR /app

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make build

RUN ls -l bin

FROM alpine:3

COPY --from=build /app/bin/php-fhir-test /php-fhir-test

ENTRYPOINT [ "/php-fhir-test" ]