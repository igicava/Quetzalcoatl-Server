FROM golang:1.22.6

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o quetzalcoatl cmd/main.go

EXPOSE 8888

CMD ["app/quetzalcoatl"]