# Use official Go image as base image
FROM golang:1.23-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod tidy

# Copy the Go source code into the container
COPY . .

# Build the Go app
RUN go build -o go-app .

# Expose the port that the Go app will run on
EXPOSE 8080

# Command to run the Go app
CMD ["./go-app"]
