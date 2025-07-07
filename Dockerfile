FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o deploytracker .

FROM scratch

COPY --from=builder /app/deploytracker /deploytracker

ENTRYPOINT ["/deploytracker"]


