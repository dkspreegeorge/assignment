FROM golang:1.20-alpine as build

WORKDIR /app

COPY . .

RUN go mod download && \
    go build -o TaskService TaskService.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/TaskService ./TaskService

EXPOSE 8089

ENTRYPOINT [ "./TaskService" ]