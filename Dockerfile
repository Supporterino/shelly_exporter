FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o shelly_exporter ./cmd/exporter

FROM scratch

COPY config.yaml /
COPY --from=build /app/shelly_exporter .

ENTRYPOINT ["/shelly_exporter", "-config", "config.yaml"]