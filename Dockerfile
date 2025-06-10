FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR /app/
COPY . .
RUN go mod download
RUN cd cmd/app; GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/convert

FROM scratch
WORKDIR /app/
COPY --from=builder /app/convert /app/convert
COPY ./.env .

ENTRYPOINT ["/app/convert"]
