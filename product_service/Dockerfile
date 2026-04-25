FROM golang:1.25 AS build

WORKDIR /product_service

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./service ./cmd/main.go

FROM scratch AS final

WORKDIR /app

COPY --from=build /product_service/service .
COPY ./migrations/* ./migrations/

EXPOSE 4000

CMD ["/app/service"]