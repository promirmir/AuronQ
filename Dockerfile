FROM golang:1.23.2-bookworm AS build
WORKDIR /src
COPY . .
RUN go test ./... && go vet ./... && CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w -buildid=" -o /out/auronq ./cmd/auronq

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/auronq /usr/local/bin/auronq
ENTRYPOINT ["/usr/local/bin/auronq"]
CMD ["version"]