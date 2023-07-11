# syntax=docker/dockerfile:1

FROM golang:1.20-alpine

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
ADD cmd ./cmd
ADD pkg ./pkg


# Install curl, tmux, bash, chromium
RUN apk update && apk upgrade  \
    && apk --no-cache add curl tmux bash chromium-swiftshader

#Create cron job
RUN echo "curl localhost:8080/stat" > stat.sh && chmod +x stat.sh
RUN echo "0 8 * * * sh /app/stat.sh" >> /var/spool/cron/crontabs/root
COPY Procfile ./
RUN go install github.com/DarthSim/overmind/v2@latest

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/main -o /docker-russian_losses_bot

EXPOSE 8080

# Run
CMD ["overmind", "s"]