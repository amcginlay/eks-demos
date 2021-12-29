FROM golang:1.12.0-alpine3.9 AS build
WORKDIR /app/
COPY main.go .
RUN go build -o main

FROM alpine:3.9
WORKDIR /app/
COPY --from=build /app/main .
RUN apk add curl bind-tools
ENV PORT=80
ENTRYPOINT [ "./main" ]