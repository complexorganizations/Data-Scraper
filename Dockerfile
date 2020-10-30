FROM golang:latest
LABEL maintainer="Prajwal Koirala <prajwalkoirala23@protonmail.com>"
WORKDIR /app
COPY main.go .
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY settings.json .
COPY scraping.json .
RUN go build
CMD ["./data-scraper"]
