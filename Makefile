GO_SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
FRONTEND_DIR = frontend

all: master slave

frontend: frontend/node_modules
	cd frontend && webpack -p

frontend/node_modules: ${FRONTEND_DIR}/package.json
	cd frontend && npm install

master: frontend vendor
	go build -o build/master github.com/arkbriar/ssmgr/master

slave: vendor
	go build -o build/slave  github.com/arkbriar/ssmgr/slave/cli

format:
	goimports -w -d ${GO_SRC}

run_master:
	build/master -c config.master.json -ca testdata/certs/ca.pem -v

run_slave:
	build/slave -c config.slave.json -v

install: install-master install-slave

install-master: master check-install
	mkdir -p /usr/local/ssmgr/bin
	cp build/master /usr/local/ssmgr/bin
	cp systemd/ssmgr.master /etc/default/
	cp systemd/ssmgr-master.service /lib/systemd/system/

install-slave: slave check-install
	mkdir -p /usr/local/ssmgr/bin
	cp build/slave /usr/local/ssmgr/bin/
	cp systemd/ssmgr.slave /etc/default/
	cp systemd/ssmgr-slave.service /lib/systemd/system/

check-install: check-linux check-systemd

check-linux:
ifneq (Linux, $(shell uname -s))
	$(error "Install is not supported on non-linux system.")
endif

check-systemd:
ifeq (, $(shell which systemctl))
	$(error "Install is not supported on system without systemd.")
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
