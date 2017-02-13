FROM golang:1.7

ADD . /go/src/github.com/arkbriar/ssmgr/

RUN go build -o /master github.com/arkbriar/ssmgr/master \
    && mkdir /data \
    && mv /go/src/github.com/arkbriar/ssmgr/config.master.json /data/config.json \
    && mv /go/src/github.com/arkbriar/ssmgr/frontend /frontend \
    && rm -rf /go/src/github.com/arkbriar/ssmgr \
    && rm -rf /frontend/node_modules

VOLUME ["/data"]

WORKDIR /data

# CA files are stored in /etc/ssmgr/ca, mounting it before running a secure master

ENTRYPOINT /master -w /frontend -ca /etc/ssmgr/ca

