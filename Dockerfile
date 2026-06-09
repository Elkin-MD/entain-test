# ---- build stage ----
FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -o /app/entaintest ./cmd

# ---- runtime stage ----
FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=build /app/entaintest /entaintest

EXPOSE 8080
ENTRYPOINT ["/entaintest"]
