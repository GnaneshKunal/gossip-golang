FROM golang as builder

WORKDIR /go/src/gossip

COPY . .

RUN go get .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /bin

COPY --from=builder /go/src/gossip/app .

ENTRYPOINT ["./app"]

EXPOSE 8080
