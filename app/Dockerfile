FROM debian:bullseye-slim
ENV TZ=Asia/Shanghai
RUN set -x && apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY ./app app

ENTRYPOINT ["/app/app"]
CMD []
