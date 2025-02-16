FROM golang:1.19
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/storage_server ./storage
EXPOSE 8081
CMD ["/app/storage_server"]
