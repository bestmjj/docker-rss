FROM golang:1.23-alpine AS build

WORKDIR /app
COPY cmd/* /app
COPY go.mod go.sum /app
RUN go mod tidy
RUN go build -o main .

FROM alpine:latest
COPY --from=build /app/main /app/main
RUN apk add tzdata
CMD ["/app/main"]
