FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the swagger doc
RUN make generate-docs

# Build natively inside the Linux container
RUN go build -o main ./cmd/server
RUN ls -l /app && file /app/main

EXPOSE 4000

CMD ["./main"]
