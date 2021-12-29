FROM golang:1.12.0-alpine3.9
ENV PORT=80
WORKDIR /app/
COPY main.go .
RUN go build -o main && \
    apk add curl bind-tools
ENTRYPOINT [ "./main" ]
