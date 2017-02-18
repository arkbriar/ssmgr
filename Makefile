GO_SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
FRONTEND_DIR = frontend
PROTOCOL_GO_SRC = protocol/master_slave.pb.go
PROTOCOL_PROTO_SRC = protocol/master_slave.proto
SLAVE_BIN = build/slave
MASTER_BIN = build/master

all: master slave

frontend: frontend/node_modules
	cd frontend && webpack -p

frontend/node_modules: ${FRONTEND_DIR}/package.json
	cd frontend && npm install

master: ${MASTER_BIN}

slave: ${SLAVE_BIN}

${MASTER_BIN}: frontend vendor ${PROTOCOL_GO_SRC}
	go build -o build/master github.com/arkbriar/ssmgr/master

${SLAVE_BIN}: vendor ${PROTOCOL_GO_SRC}
	go build -o build/slave  github.com/arkbriar/ssmgr/slave/cli

${PROTOCOL_GO_SRC}: ${PROTOCOL_PROTO_SRC}
	go generate

format:
	goimports -w -d ${GO_SRC}

check-debugger:
ifeq (, $(shell which dlv))
	$(error Debugger dlv is not found)
endif

copy-config-files:
	@ if [ ! -e build/config.master.json ]; then cp config.master.json build/; fi
	@ if [ ! -e build/config.slave.json ]; then cp config.slave.json build/; fi

run-master: copy-config-files ${MASTER_BIN}
	build/master -c build/config.master.json -ca testdata/certs/ca.pem

run-slave: copy-config-files ${SLAVE_BIN}
	build/slave -c build/config.slave.json

debug-master: check-debugger copy-config-files ${MASTER_BIN}
	dlv exec build/master -- -c build/config.master.json -ca testdata/certs/ca.pem -v

debug-slave: check-debugger copy-config-files ${SLAVE_BIN}
	dlv exec build/slave -- -c build/config.slave.json -v

install: install-master install-slave

install-master: check-install ${MASTER_BIN}
	mkdir -p /usr/local/ssmgr/bin
	mkdir -p /etc/ssmgr
	cp config.master.json /etc/ssmgr/
	cp -R frontend /etc/ssmgr/frontend
	cp build/master /usr/local/ssmgr/bin/
	cp systemd/ssmgr.master /etc/default/
	cp systemd/ssmgr-master.service /lib/systemd/system/

install-slave: check-install ${SLAVE_BIN}
	mkdir -p /usr/local/ssmgr/bin
	mkdir -p /etc/ssmgr
	cp config.slave.json /etc/ssmgr/
	cp build/slave /usr/local/ssmgr/bin/
	cp systemd/ssmgr.slave /etc/default/
	cp systemd/ssmgr-slave.service /lib/systemd/system/

check-install: check-linux check-systemd

check-linux:
ifneq (Linux, $(shell uname -s))
	$(error Install is not supported on non-linux system)
endif

check-systemd:
ifeq (, $(shell which systemctl))
	$(error Install is not supported on system without systemd)
endif

docker:
	docker build . --no-cache -t ssmgr-master

.PHONY: all frontend master slave format docker check-install install clean

vendor: glide.lock glide.yaml
	glide install
	go install github.com/arkbriar/ssmgr/vendor/github.com/mattn/go-sqlite3

clean:
	rm -f ssmgr.db
	rm -rf build
