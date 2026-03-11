ARG SOURCE=build

FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X 'localgate/internal/version.Version=${VERSION}' \
              -X 'localgate/internal/version.Commit=${COMMIT}' \
              -X 'localgate/internal/version.BuildDate=${BUILD_DATE}'" \
    -o /localgate .

FROM scratch AS source-build
COPY --from=builder /localgate /localgate

FROM scratch AS source-prebuilt
ARG TARGETARCH=amd64
COPY localgate-linux-${TARGETARCH} /localgate

FROM source-${SOURCE} AS binary

FROM alpine:3

COPY --from=binary /localgate /usr/local/bin/localgate
COPY entrypoint.sh /entrypoint.sh
RUN <<_EOF_
chmod +x /usr/local/bin/localgate
chmod +x /entrypoint.sh
_EOF_

EXPOSE 9000

ENTRYPOINT ["/entrypoint.sh"]
