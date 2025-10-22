FROM golang:1.24 AS builder
WORKDIR /app
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/insider ./cmd/main.go
FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /bin/insider /bin/insider
COPY internal/docs/swagger.yaml /internal/docs/swagger.yaml
ENV HTTP_PORT=8080
EXPOSE 8080
ENTRYPOINT ["/bin/insider"]
