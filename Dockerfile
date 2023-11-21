# Build stage
FROM golang:1.20.2-alpine3.16 AS build-env
RUN apk --no-cache add build-base gcc
ADD . /build
WORKDIR /build
RUN make build


# Final stage
FROM alpine:3.18.4

WORKDIR /app
COPY --from=build-env /build/dropbox-auto-cleaner /app/

ENV DROPBOX_CLEANER_API_TOKEN=""

ENTRYPOINT ["./dropbox-auto-cleaner