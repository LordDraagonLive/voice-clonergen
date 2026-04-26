FROM golang:1.26-alpine AS build
WORKDIR /src
COPY apps/api/go.mod apps/api/go.mod
WORKDIR /src/apps/api
RUN go mod download
WORKDIR /src
COPY apps/api apps/api
WORKDIR /src/apps/api
RUN go build -o /out/api ./cmd/server

FROM alpine:3.22
WORKDIR /app
COPY --from=build /out/api /app/api
EXPOSE 8080
CMD ["/app/api"]
