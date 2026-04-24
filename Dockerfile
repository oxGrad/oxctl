# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/oxctl ./cmd/oxctl

FROM python:3.12-alpine AS final
RUN pip install --no-cache-dir awscli

COPY --from=builder /bin/oxctl /usr/local/bin/oxctl

ENTRYPOINT ["oxctl"]
