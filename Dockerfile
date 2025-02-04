# use the official Golang image as a base: latest stable version
FROM golang:1.22-alpine

# set the working directory
WORKDIR /app

# copy the Go module files
COPY go.mod ./

# copy the source code
COPY . .

# build the application
RUN go build -o kubernetes-api .

# expose the port the app runs on
EXPOSE 8080

# run the application
CMD ["./kubernetes-api"]