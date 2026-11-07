FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.* ./
COPY . .
# go mod tidy creates go.sum on the first build; commit go.sum once it exists.
RUN go mod tidy && CGO_ENABLED=0 go build -o /out/goodbuy ./cmd/server

FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
COPY --from=build /out/goodbuy /usr/local/bin/goodbuy
EXPOSE 8080
ENTRYPOINT ["goodbuy"]
