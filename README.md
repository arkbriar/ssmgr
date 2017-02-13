# SSMGR

Shadowsocks manager of multiple servers, providing simple and easy way for users to acquire shadowsocks services over regions of servers.

**NOTIFICATION** *After [commit ca33594](https://github.com/arkbriar/ssmgr/pull/23/commits/ca335940389f4a9ec937386a898880d52b529f70), slave's managed dir changes to $HOME/.ssmgr, refer to [#23](https://github.com/arkbriar/ssmgr/pull/23) if you want to do a migration.*

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

**NOTE**, `install` is only supported on linux with systemd.

Before install ssmgr, install shadowsocks-libev first.

```
sudo apt install shadowsocks-libev
```

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
| ssmgr slave  |    | ssmgr slave  |  ...  |       |
+--------------+    +--------------+       +-------+
       |                    |                  |
       +------------+-------+-------  ...  ----+
                    |
                    |
             +---------------+
             | ssmgr master  |
             +---------------+
```

We define the protocol as rpc methods in [a .proto file](protocol/master_slave.proto).

## Known Issues

1. [Issues](https://github.com/arkbriar/ssmgr/issues?q=is%3Aopen+is%3Aissue+label%3Abug) here with `bug` tags.

## License

Licensed under [GPL v3](LICENSE).


