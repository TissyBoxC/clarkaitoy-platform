FROM golang:1.27 AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/device-platform ./cmd/device-platform

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/device-platform /device-platform
EXPOSE 8081
ENTRYPOINT ["/device-platform"]
