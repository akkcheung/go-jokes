FROM golang:1.18-alpine3.16 AS builder
RUN apk add build-base
RUN mkdir /build
ADD go.mod go.sum main.go /build/
WORKDIR /build
RUN go build

FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/go-jokes /app/ 
COPY views/ /app/views
WORKDIR /app
CMD ["./go-jokes"]
