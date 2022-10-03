FROM public.ecr.aws/docker/library/golang:1.19 as build
WORKDIR /src

COPY . .
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -buildvcs=false -o kubernetes-diff-logger .

FROM gcr.io/distroless/static:nonroot
WORKDIR /
USER 65532:65532

COPY --from=build /src/kubernetes-diff-logger .

ENTRYPOINT ["/kubernetes-diff-logger"]
