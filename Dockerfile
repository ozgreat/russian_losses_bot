FROM golang:1.22.5 AS builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY cmd ./cmd
COPY pkg ./pkg

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/main -o /russian_losses_bot

FROM alpine:latest AS runner

RUN apk --no-cache add ca-certificates mailcap && addgroup -S app && adduser -S app -G app
USER app

WORKDIR /app
COPY --from=builder /russian_losses_bot .

EXPOSE 8080

# Run
ENTRYPOINT ["./russian_losses_bot"]
