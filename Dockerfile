FROM golang:1.20 AS builder

WORKDIR /workspace
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o webhook main.go

FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/webhook webhook

USER nonroot:nonroot

ENTRYPOINT ["/webhook"]
