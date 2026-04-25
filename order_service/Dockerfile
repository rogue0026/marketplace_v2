FROM golang:1.25 AS build

WORKDIR /order_service

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o service ./cmd/main.go

FROM scratch AS final

WORKDIR /app

COPY --from=build /order_service/service ./
COPY ./migrations/ ./migrations/

EXPOSE 4000

CMD ["/app/service"]