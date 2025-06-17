# syntax=docker/dockerfile:1.4

# Build the manager binary
FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/golang:1.19 as builder

ARG TARGETARCH
ARG TARGETOS
WORKDIR /src

# Copy the go source
COPY . .

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -a -buildvcs=false -o kubernetes-diff-logger .


FROM gcr.io/distroless/static:nonroot
WORKDIR /
USER 65532:65532

COPY --from=build /src/kubernetes-diff-logger .

ENTRYPOINT ["/kubernetes-diff-logger"]
