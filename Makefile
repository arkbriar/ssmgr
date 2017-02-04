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
