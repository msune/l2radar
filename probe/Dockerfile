FROM golang:1.24-bookworm AS build

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        clang \
        llvm \
        libbpf-dev \
        linux-libc-dev \
        gcc-multilib \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go generate ./pkg/loader/
RUN CGO_ENABLED=0 go build -o /l2radar ./cmd/l2radar/

FROM debian:bookworm-slim
COPY --from=build /l2radar /l2radar
ENTRYPOINT ["/l2radar"]
