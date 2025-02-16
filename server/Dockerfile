FROM golang:1.19

WORKDIR /app

# Установка MySQL клиента
RUN apt-get update && apt-get install -y default-mysql-client && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main ./server

COPY wait-for-mysql.sh /wait-for-mysql.sh
RUN chmod +x /wait-for-mysql.sh

EXPOSE 8080
CMD ["/wait-for-mysql.sh", "./main"]
