FROM alpine:3

ARG TARGETARCH=amd64

COPY localgate-linux-${TARGETARCH} /usr/local/bin/localgate
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 9000

ENTRYPOINT ["/entrypoint.sh"]
