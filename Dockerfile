FROM golang:1.22.3 as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o main .
FROM chromedp/headless-shell:latest
RUN mkdir /app
COPY --from=builder /app/main /app/main

EXPOSE 8080

ENTRYPOINT ["/app/main"]