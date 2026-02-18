FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /nexusclaw ./cmd/nexusclaw

FROM gcr.io/distroless/static-debian12
COPY --from=builder /nexusclaw /nexusclaw
EXPOSE 8080
ENTRYPOINT ["/nexusclaw"]
CMD ["serve"]
