
# We must use a glibc based distro due to embedded python not supporting musl libc for aarch64
FROM debian:bullseye-slim
COPY manager /manager
USER 65532:65532

ENTRYPOINT ["/manager"]
