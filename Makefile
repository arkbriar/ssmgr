GO_SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

all: frontend server slave

frontend: frontend/node_modules
	cd frontend && webpack -p

master: frontend vendor
	go build -o build/master github.com/arkbriar/ss-mgr/master

slave: vendor
	go build -o build/slave  github.com/arkbriar/ss-mgr/slave/cli

format:
	goimports -w $(GO_SRC)

run_master:
	build/master -c config.master.json -v

run_slave:
	build/slave -c config.slave.json -v

install: install-master install-slave

install-master: master
	mkdir -p /usr/local/ssmgr/bin
	mv build/master /usr/local/ssmgr/bin
	mv systemd/ssmgr.master /etc/default/
	mv systemd/ssmgr-master.service /lib/systemd/system/

install-slave: slave check-install
	mkdir -p /usr/local/ssmgr/bin
	mv build/slave /usr/local/ssmgr/bin/
	mv systemd/ssmgr.slave /etc/default/
	mv systemd/ssmgr-slave.service /lib/systemd/system/

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

.PHONY: all frontend server slave format docker check-install

vendor: glide.lock glide.yaml
	glide install
	go install github.com/arkbriar/ss-mgr/vendor/github.com/mattn/go-sqlite3

frontend/node_modules:
	cd frontend && npm install

clean:
	rm -f ssmgr.db
	rm -rf build
