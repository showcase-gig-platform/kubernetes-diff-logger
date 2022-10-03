FROM golang:1.19 as build
WORKDIR /src

COPY . .
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -o app .

FROM gcr.io/distroless/static:nonroot
WORKDIR /
USER 65532:65532

COPY --from=build /src/app .

ENTRYPOINT ["/app"]
