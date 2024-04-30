FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 GOOS=linux go build -o api main.go

FROM golang:1.22-alpine

WORKDIR /app

COPY --from=builder /app/api ./

EXPOSE 8000

CMD [ "./api" ]