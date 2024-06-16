ARG PHP_VERSION

FROM ghcr.io/haokeyingxiao/haoke-cli-base:${PHP_VERSION}

COPY haoke-cli /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/haoke-cli"]
CMD ["--help"]
