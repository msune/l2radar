FROM golang:1.24-bookworm AS build

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        clang \
        llvm \
        libbpf-dev \
        linux-libc-dev \
    && if [ "$(dpkg --print-architecture)" = "amd64" ]; then \
        apt-get install -y --no-install-recommends gcc-multilib; \
    else \
        for d in /usr/include/$(uname -m)-linux-gnu/*/; do \
            [ ! -e /usr/include/$(basename "$d") ] && \
                ln -s "$d" /usr/include/$(basename "$d"); \
        done; \
    fi \
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
