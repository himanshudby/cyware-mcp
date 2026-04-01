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

ENV PORT=5421

EXPOSE 5421

ENTRYPOINT ["/app/out"]
CMD ["-config_path", "/app/cmd/config.remote.yaml"]

