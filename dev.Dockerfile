FROM golang:1.24

WORKDIR /app

# Install air
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

COPY ./cmd/server/.env ./cmd/server/.env

EXPOSE 4000

CMD ["air", "sh", "-c", "make generate-docs && air -c .air.toml"]
