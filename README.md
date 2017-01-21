# ss-mgr

Shadowsocks manager of multiple servers.

## Compile

## Install

## Docs

## Develop

### Schedule

Tasks | Original Due | Progress | Realized Date
:-: | :-: | :-: | :-:
Manager protocol and implementation | 2017-01-24 | `[==================90%]` | 
Persistense layer (for master) | 2017-02-01 | `[1%===================]` |
Slave server management (process monitor/manager) | 2017-02-05 | `[0%===================]` |
Plugin protocol and implementation | 2017-02-10 | `[0%===================]` |
Advanced features `*` | 2017-02-28 | `[0%===================]` |

P.S. tasks with `*` are not considered in release plans.

1. We will freeze feature when tasks without `*` are all realized.
2. When developing advanced features, we will provide an option to disable them.


### Protocols

There are two roles in manager system, **master** and **slave**. Master manages all slaves where shadowsocks manager/server runs.

There are two sets of protocols, one is called manager protocol and another is plugin protocol. 

#### Manager Protocol

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

#### Plugin Protocol

## License

Licensed under [GPL v3](LICENSE).


