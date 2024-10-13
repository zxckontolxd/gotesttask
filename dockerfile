FROM golang:1.19 as builder
WORKDIR /app
COPY . .
RUN go mod init wallet-app
RUN go get -d -v ./...
RUN go build -o wallet-app

FROM gcr.io/distroless/base-debian10
COPY --from=builder /app/wallet-app /wallet-app
CMD ["/wallet-app"]