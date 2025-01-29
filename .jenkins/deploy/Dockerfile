FROM mcr.microsoft.com/oss/go/microsoft/golang:1.22-fips-bullseye AS conjur-cli-go-builder

ENV VERSION=""
ARG ARCH="arm64"

COPY . /github.com/cyberark/conjur-cli-go

WORKDIR /github.com/cyberark/conjur-cli-go

RUN chmod +x .jenkins/deploy/builder_entrypoint.sh
RUN ./.jenkins/deploy/builder_entrypoint.sh

FROM alpine:3.21 AS conjur-cli-go

# Required to add dynamically linked libraries for conjur-cli-go
RUN apk add libc6-compat

COPY --from=conjur-cli-go-builder /github.com/cyberark/conjur-cli-go/dist/conjur /usr/local/bin/conjur

ENTRYPOINT ["/usr/local/bin/conjur"]
