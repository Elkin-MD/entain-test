# ---- build stage ----
FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/migrate ./cmd/migrate

# ---- runtime stage ----
FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=build /app/server /server
COPY --from=build /app/migrate /migrate

EXPOSE 8080
ENTRYPOINT ["/server"]
