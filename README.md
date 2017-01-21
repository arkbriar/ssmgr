# ss-mgr

Shadowsocks manager of multiple servers.

## Compile

## Install

## Docs

## Develop

There are two roles in manager system, **master** and **slave**. Master manages all slaves where shadowsocks manager/server runs.

There are two sets of protocols, one is called manager protocol and another is plugin protocol. 

### Manager Protocol

Manager protocol is designed for communication between slaves and the master.

```
+--------------+    +--------------+       +-------+
|  ss-manager  |    |  ss-manager  |  ...  |       |
+--------------+    +--------------+       +-------+
        |                   |                  |
        |                   |                  |
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

We define the protocol as rpc methods in [a .proto file](manager/protocol/shadowsocks_manager.proto).

### Plugin Protocol

## License

Licensed under [GPL v3](LICENSE).


