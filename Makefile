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

# Support install command on linux with systemd.
UNAME_S = $(shell uname -s)
SYSTEMCTL = $(shell command -v systemctl)
LINUX_SYSTEMD = 0
ifeq (${UNAME_S}, "Linux")
ifneq (${SYSTECTL}, "")
	LINUX_SYSTEMD = 1
endif
endif

install: install-master install-slave

install-master: master
	@ if [ ${LINUX_SYSTEMD} -eq 1 ]; then \
		echo "installing binaries, env file and systemd unit of ssmgr master" && \
		mkdir -p /usr/local/ssmgr/bin && \
		mv build/master /usr/local/ssmgr/bin/ && \
		mv systemd/ssmgr.master /etc/default/ && \
		mv systemd/ssmgr-master.service /lib/systemd/system/ \
		; \
		else \
		echo "Not supported" \
		; \
		fi

install-slave: slave
	@ if [ ${LINUX_SYSTEMD} -eq 1 ]; then \
		echo "installing binaries, env file and systemd unit of ssmgr slave" && \
		mkdir -p /usr/local/ssmgr/bin && \
		mv build/slave /usr/local/ssmgr/bin/ && \
		mv systemd/ssmgr.slave /etc/default/ && \
		mv systemd/ssmgr-slave.service /lib/systemd/system/ \
		; \
		else \
		echo "Not supported" \
		; \
		fi \


docker:
	docker build . --no-cache -t ssmgr-master

.PHONY: all frontend server slave format docker

vendor: glide.lock glide.yaml
	glide install
	go install github.com/arkbriar/ss-mgr/vendor/github.com/mattn/go-sqlite3

frontend/node_modules:
	cd frontend && npm install

clean:
	rm -f ssmgr.db
	rm -rf build
