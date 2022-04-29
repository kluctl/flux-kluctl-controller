# Build the manager binary
FROM golang:1.18-bullseye as builder

WORKDIR /workspace

# need pip3 for the generate phase
RUN apt update && apt install python3 python3-pip unzip -y

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# generate
RUN go mod vendor
RUN cd vendor/github.com/kluctl/kluctl/v2/pkg/jinja2 && go run ./generate
RUN cd vendor/github.com/kluctl/kluctl/v2/pkg/python && go run ./generate

# Build
RUN CGO_ENABLED=0 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/manager /manager
USER 65532:65532

ENTRYPOINT ["/manager"]
