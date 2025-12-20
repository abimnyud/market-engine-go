FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /market-engine ./cmd/market-engine/main.go

FROM gcr.io/distroless/static-debian11 AS final

WORKDIR /app

COPY --from=build /market-engine /app/market-engine

EXPOSE 50051

ENTRYPOINT ["/app/market-engine"]
