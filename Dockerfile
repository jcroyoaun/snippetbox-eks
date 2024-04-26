# jcroyoaun/snippetbox:1.2
# Build stage
FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/web

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=build /app/main .
COPY --from=build /app/ui/html /app/ui/html
COPY --from=build /app/ui/static /app/ui/static

EXPOSE 4000

CMD ["./main"]
