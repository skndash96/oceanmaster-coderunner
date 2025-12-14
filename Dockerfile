# syntax=docker/dockerfile:1.5.2

# Build nsjail
FROM debian:bookworm-20230703-slim AS nsjail
WORKDIR /src
RUN apt-get update && \
    apt-get install -y autoconf bison flex gcc g++ libnl-route-3-dev libprotobuf-dev libseccomp-dev libtool make pkg-config protobuf-compiler
COPY nsjail .
RUN make -j

# Build python mount
FROM python:3.12-slim AS jail
RUN pip install --no-cache-dir numpy && \
    mkdir /submission && \
    touch /submission/main.py
COPY wrapper.py /

# Build go runner
FROM golang:1.24.0-bookworm AS run
WORKDIR /src
RUN apt-get update && apt-get install -y libseccomp-dev libgmp-dev
COPY go.mod go.sum ./
RUN go mod download
COPY cmd cmd
COPY internal internal
RUN go build -v -ldflags '-w -s' ./cmd/runner

# Final Image
FROM busybox:1.36.1-glibc
RUN mkdir -p /app/cgroup/unified /submissions
COPY --link --from=nsjail /usr/lib/*-linux-gnu/libprotobuf.so.32 /usr/lib/*-linux-gnu/libnl-route-3.so.200 \
    /lib/*-linux-gnu/libnl-3.so.200 /lib/*-linux-gnu/libz.so.1 /usr/lib/*-linux-gnu/libstdc++.so.6 \
    /lib/*-linux-gnu/libgcc_s.so.1 /lib/
COPY --link --from=run /usr/lib/*-linux-gnu/libseccomp.so.2 /usr/lib/*-linux-gnu/libgmp.so.10 /lib/
COPY --link --from=run /src/runner /app/runner
COPY --link --from=nsjail /src/nsjail /app/nsjail
COPY --from=jail / /srv
CMD ["/app/runner"]
