ARG GO_VERSION
ARG ALPINE_VERSION
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as build
WORKDIR /src
COPY . .
RUN go install .

FROM alpine:${ALPINE_VERSION}
COPY --from=build /go/bin/transmission-nat-pmp /usr/bin/transmission-nat-pmp
ENTRYPOINT ["transmission-nat-pmp"]
