FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/oxctl ./cmd/oxctl

FROM amazon/aws-cli:2.34.38 AS release

COPY --from=builder /bin/oxctl /usr/local/bin/oxctl

ENTRYPOINT ["oxctl"]
