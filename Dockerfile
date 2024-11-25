FROM golang:1.21.5-alpine as base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM golang:1.21.5-alpine as final

WORKDIR /app

COPY --from=base /app/main ./main
COPY --from=base /app/config ./config/

EXPOSE 8080
CMD ["./main"]
