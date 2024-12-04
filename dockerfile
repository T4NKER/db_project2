# Use the Go base image
FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o db_project2

# Expose the application port
EXPOSE 3000

# Start the application
CMD ["./db_project2"]