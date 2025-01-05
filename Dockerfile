FROM golang:latest as builder
WORKDIR /app
ADD . /app/
RUN CGO_ENABLED=0 go build -o wallet .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/wallet .
EXPOSE 8080
ENTRYPOINT ["./wallet"]
