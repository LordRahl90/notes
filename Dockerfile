FROM golang:latest as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o notes ./cmd


FROM scratch

COPY --from=builder /app/notes /notes

EXPOSE 8181
# Run
CMD ["/notes"]