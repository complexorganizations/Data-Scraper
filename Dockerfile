FROM golang:latest
LABEL maintainer="Prajwal Koirala <prajwalkoirala23@protonmail.com>"
WORKDIR /app
COPY main.go .
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY output.json .
COPY scraping.json .
COPY settings.json .
RUN go build
CMD ["./Data-Scraper"]
