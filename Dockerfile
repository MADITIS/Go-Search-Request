FROM golang:1.20.3-bullseye

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download



COPY . .
RUN ls -la /app


RUN go build -o main ./src

EXPOSE 8080

CMD ["/app/main"]