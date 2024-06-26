FROM golang:1.20-alpine AS build

RUN apk add build-base git linux-headers libc-dev
RUN go install github.com/go-delve/delve/cmd/dlv@latest

WORKDIR /work
COPY go.mod go.sum /work/
COPY errors/go.mod errors/go.sum /work/errors/
COPY math/go.mod math/go.sum /work/math/
COPY api/go.mod api/go.sum /work/api/
COPY core/go.mod core/go.sum /work/core/
COPY depinject/go.mod depinject/go.sum /work/depinject/
RUN go mod download

ENV CGO_CFLAGS="-O -D__BLST_PORTABLE__"
ENV CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"

COPY ./ /work
RUN LEDGER_ENABLED=false make COSMOS_BUILD_OPTIONS="debug,nostrip" clean build


FROM alpine AS run
RUN apk add bash curl jq libstdc++
EXPOSE 26656 26657
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start", "--log_format", "plain"]
STOPSIGNAL SIGTERM
VOLUME /simd
WORKDIR /simd

COPY contrib/images/simd-dlv/wrapper.sh /usr/bin/wrapper.sh
COPY --from=build /work/build/simd /simd/
COPY --from=build /go/bin/dlv /usr/local/bin
