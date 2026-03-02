# syntax=docker/dockerfile:1

FROM golang:1.22 AS builder
WORKDIR /src

COPY go.mod .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/ispnag ./cmd/ispnag

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=builder /out/ispnag /ispnag
USER nonroot:nonroot
ENTRYPOINT ["/ispnag"]
