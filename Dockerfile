FROM golang:latest
LABEL maintainer="Prajwal Koirala <prajwalkoirala23@protonmail.com>"
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build
CMD ["./Data-Scraper"]
