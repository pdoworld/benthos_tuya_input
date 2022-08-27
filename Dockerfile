FROM golang:1.17 AS build

RUN useradd -u 10001 benthos

WORKDIR /build/
COPY . /build/

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor

FROM busybox AS package

LABEL maintainer="Ashley Jeffs <ash@jeffail.uk>"

WORKDIR /

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /build/benthos_tuya_input .
COPY ./config/example.yaml /benthos.yaml

RUN mkdir /logs
# RUN chown benthos:10001 benthos_tuya_input
RUN chown -R benthos:10001 /logs

USER benthos

EXPOSE 4195

ENTRYPOINT ["/benthos_tuya_input"]

CMD ["-c", "/benthos.yaml"]
