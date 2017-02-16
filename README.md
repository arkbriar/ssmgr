# SSMGR

Shadowsocks manager of multiple servers, providing simple and easy way for users to acquire shadowsocks services over regions of servers and advanced tools for operators to manage the servers.

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

**[NOTE]** `install` is only supported on linux with systemd.
**[NOTE]** If you want to enable TLS, you should modify the env and config files and specify your certificates. For more details about generating self-signed certificates, please see [Generate Self-signed Certificates](#generate-self-signed-certificates).

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

### Enable TLS

Enable TLS to secure the communication between master and slaves.

**TLS on Slave**

Add "tls" field to config.json file.

```json
{
  "port": 6001,
  "manager_port": 6001,
  "token": "SSMGRTEST",
  "tls": {
    "cert_file": "testdata/certs/server.crt",
    "key_file": "testdata/certs/server.key"
  }
}
```

**TLS on Master**

Specify CA X.509 file when you start the master.

```bash
master -w frontend -c config.json -ca path/to/ca.pem
```

### Generate Self-signed Certificates

Generate CA key and PEM file if you do not have one:

```
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.pem
```

Then, using script `gssc` to generate these certificates.

```bash
./tools/gssc --ip SLAVE_IP --ca CA_DIR
```

This will generate a 2048 bits key and a certificate whose expiry is 365 days in current directory.

For more details of `gssc`, please see

```bash
./tools/gssc -h
```

### Log to Slack

We implement a hook of logrus to send some levels of logs to slack channel. This helps developers to monitor servers and to develop ChatOps in the future.

Slack logs is only supported on master. To enable it, add the "slack" field to config.json file, 

```json
{
  "...": "...",
  "slack": {
    "channel": "#SLACK_CHANNEL",
    "token": "TEST_SLACK_TOKEN",
    "levels": [
      "panic",
      "fatal",
      "error",
      "warn",
      "info",
      "debug"
    ]
  }
}
```

where token is the application token and levels are the logrus levels to send.

## Known Issues

1. [Issues](https://github.com/arkbriar/ssmgr/issues?q=is%3Aopen+is%3Aissue+label%3Abug) here with `bug` tags.

## License

Licensed under [GPL v3](LICENSE).


