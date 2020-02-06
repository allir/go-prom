FROM golang:1.13 as build
WORKDIR /build
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o go-prom -ldflags="-s -w"


FROM scratch
EXPOSE 8080
COPY --from=build /build/go-prom /go-prom
CMD ["/go-prom"]
