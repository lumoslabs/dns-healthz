FROM instrumentisto/glide:0.13.0-go1.9
WORKDIR /go/src/github.com/lumoslabs/dns-healthz
RUN DUMB_INIT_VERSION=1.2.1 \
    && apk add --update --no-cache ca-certificates openssl \
    && wget --no-check-certificate -O /dumb-init https://github.com/Yelp/dumb-init/releases/download/v${DUMB_INIT_VERSION}/dumb-init_${DUMB_INIT_VERSION}_amd64 \
    && chmod -v +x /dumb-init
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
COPY ["glide*", "vendor", "./"]
RUN glide install --strip-vendor
COPY . ./
RUN cd cmd/dns-healthz \
    && go build -v \
    -ldflags "-s" \
    -installsuffix cgo -o /dns-healthz . \
    && chmod -v +x /dns-healthz

FROM scratch
ENTRYPOINT ["/dumb-init", "--", "/dns-healthz"]
COPY --from=0 /dumb-init .
COPY --from=0 /dns-healthz .
