# SSMGR

Shadowsocks manager of multiple servers, providing simple and easy way for users to acquire shadowsocks services over regions of servers.

## Compile

First of all, install the dependent tools.

For mac users,

```bash
brew install go node glide
npm install -g webpack
```

For ubuntu/debian users, install golang toolchain (>=1.7) and

```bash
sudo apt install nodejs-legacy npm 
sudo npm install -g webpack
```

Then,

```bash
make all
```

## Install

To install it completely (binaries and systemd services), 

```bash
# Install to /usr/local/ssmgr/*, /etc/default/ssmgr.{master/slave} and /lib/systemd/system/ssmgr-{master/slave}.service
sudo make install
```

Or you can install master/slave seperately,

```bash
# Master
sudo make install-master
# Slave
sudo make install-slave
```

## Docs

### Protocols

#### Master-Slave

Master and slaves are organized as: 

```
+--------------+    +--------------+       +-------+
| ss-mgr slave |    | ss-mgr slave |  ...  |       |
+--------------+    +--------------+       +-------+
       |                    |                  |
       +------------+-------+-------  ...  ----+
                    |
                    |
             +---------------+
             | ss-mgr master |
             +---------------+
```

We define the protocol as rpc methods in [a .proto file](manager/protocol/master_salve.proto).

## Known Issues

1. [Issues](https://github.com/arkbriar/ss-mgr/issues) here with `bug`, `enhancement` and other tags.

## License

Licensed under [GPL v3](LICENSE).


