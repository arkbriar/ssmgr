FROM golang:1.7

ADD . /go/src/github.com/arkbriar/ss-mgr/

RUN go build -o /master github.com/arkbriar/ss-mgr/master \
    && mkdir /data \
    && mv /go/src/github.com/arkbriar/ss-mgr/config.json /data/ \
    && mv /go/src/github.com/arkbriar/ss-mgr/frontend /frontend \
    && rm -rf /go/src/github.com/arkbriar/ss-mgr \
    && rm -rf /frontend/node_modules

VOLUME ["/data"]

WORKDIR /data

ENTRYPOINT /master -w /frontend

