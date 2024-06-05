ARG BUILDIMAGE=golang:1.22.4
ARG BASEIMAGE=alpine:3.20.0

FROM ${BUILDIMAGE} as builder
WORKDIR /src
ADD . /src
RUN CGO_ENABLED=0 go build -o witness-webhook -ldflags '-s -d -w' ./main.go

FROM ${BASEIMAGE}
COPY --from=builder /src/witness-webhook /witness-webhook
ENTRYPOINT ["/witness-webhook"]
EXPOSE 8085/tcp
