FROM mcr.microsoft.com/oss/go/microsoft/golang:1.22-fips-bullseye AS conjur-cli-go-builder

ENV VERSION=""
ENV GO111MODULE='auto'

COPY ./bin/builder_entrypoint.sh ./builder_entrypoint.sh
RUN chmod +x ./builder_entrypoint.sh

COPY . .

RUN ls -la

RUN ./builder_entrypoint.sh

FROM alpine:3.21 AS conjur-cli-go

COPY --from=conjur-cli-go-builder /usr/local/bin/conjur /usr/local/bin/conjur

ENTRYPOINT ["/usr/local/bin/conjur"]
