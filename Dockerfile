FROM golang:1.12-alpine as builder

RUN apk add git

RUN mkdir -p /usr/src/psp-validating-admission-webhook/tmp

ADD ./go.mod ./go.sum /usr/src/psp-validating-admission-webhook/
WORKDIR /usr/src/psp-validating-admission-webhook

RUN go mod download

ADD . /usr/src/psp-validating-admission-webhook

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o webhookserver .

FROM scratch
COPY --from=builder /usr/src/psp-validating-admission-webhook/webhookserver /psp-validating-admission-webhook/webhookserver

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder --chown=nobody:nobody /usr/src/psp-validating-admission-webhook/tmp /tmp
USER nobody

WORKDIR /psp-validating-admission-webhook
ENTRYPOINT ["./webhookserver"]