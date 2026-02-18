FROM node:20-alpine AS frontend

WORKDIR /app/web
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend /app/web/dist ./web/dist

RUN CGO_ENABLED=0 go build -o server ./cmd/server

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/server .

RUN mkdir -p /app/data/uploads /app/data/matches

EXPOSE 3001

CMD ["./server"]
