FROM golang:1.24.2 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out ./cmd

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app

COPY --from=build /out /app/out
COPY --from=build /app/cmd/config.remote.yaml /app/cmd/config.remote.yaml

# Railway commonly routes to port 8080 for Docker deployments.
# The app will still honor $PORT if Railway sets a different value.
ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["/app/out"]
CMD ["-config_path", "/app/cmd/config.remote.yaml"]

