FROM alpine:3.3

ENV MICROMDM_VERSION=v0.1.0.1-dev
RUN apk --no-cache add curl && \
    curl -L https://github.com/micromdm/micromdm/releases/download/${MICROMDM_VERSION}/micromdm-linux-amd64 -o /micromdm && \
    chmod a+x /micromdm && \
    apk del curl

CMD ["/micromdm"]
