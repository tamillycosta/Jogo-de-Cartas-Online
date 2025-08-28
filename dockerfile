FROM golang:1.25

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Builda o binário chamado 'app'
RUN go build -o app .

EXPOSE 8080
CMD ["./app"]
