FROM golang:1.17.2-alpine3.14 as build
ENV GOOS linux
ENV CGO_ENABLED 0
RUN mkdir /app
COPY /go.mod /app/go.mod
COPY /internal /app/internal
COPY /cmd/restapi/main.go /app/main.go
WORKDIR /app
RUN go mod tidy
RUN go build -o restapi main.go

FROM alpine:3.14 as deploy
ARG DATABASE_DSN
ARG KAFKA_DSN
ARG KAFKA_TOPIC
ARG KAFKA_PARTITION
ARG APP_PORT
ENV DATABASE_DSN=$DATABASE_DSN
ENV KAFKA_DSN=$KAFKA_DSN
ENV KAFKA_TOPIC=$KAFKA_TOPIC
ENV KAFKA_PARTITION=$KAFKA_PARTITION
ENV APP_PORT=$APP_PORT
RUN apk add --no-cache curl
RUN apk add --no-cache ca-certificates
COPY --from=build /app/restapi /app/restapi
CMD cd /app/ && ./restapi
