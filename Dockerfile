FROM golang:1.10.3-alpine as build-env
WORKDIR /go/src/github.com/marcusolsson/goddd/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o goapp ./cmd/shippingsvc

FROM alpine:3.7
WORKDIR /app
COPY --from=build-env /go/src/github.com/marcusolsson/goddd/booking/docs ./booking/docs
COPY --from=build-env /go/src/github.com/marcusolsson/goddd/tracking/docs ./tracking/docs
COPY --from=build-env /go/src/github.com/marcusolsson/goddd/handling/docs ./handling/docs
COPY --from=build-env /go/src/github.com/marcusolsson/goddd/goapp .
EXPOSE 8080
ENTRYPOINT ["./goapp"]
