FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /pr-checklist-collector .

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

COPY --from=builder /pr-checklist-collector /usr/local/bin/pr-checklist-collector

ENTRYPOINT ["pr-checklist-collector"]
