FROM golang:1.27 AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/voice-gateway ./cmd/voice-gateway

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/voice-gateway /voice-gateway
EXPOSE 8082
ENTRYPOINT ["/voice-gateway"]
