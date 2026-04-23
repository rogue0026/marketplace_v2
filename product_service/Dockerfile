FROM golang:1.25 AS build

WORKDIR /product_service

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./service ./cmd/main.go

FROM scratch AS final

COPY --from=build /product_service/service /

EXPOSE 8080

CMD ["/service"]