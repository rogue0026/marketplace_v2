FROM golang:1.25 AS build

WORKDIR /gateway

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o service ./cmd/main.go

FROM scratch AS final

WORKDIR /app

COPY --from=build /gateway/service ./service

EXPOSE 8080

CMD ["/app/service"]