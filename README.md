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
Persistence layer (for master) | 2017-02-01 | `[==10%================]` |
Slave server management (process monitor/manager) | 2017-02-05 | `[0%===================]` |
Plugin protocol and implementation | 2017-02-10 | `[0%===================]` |
Advanced features `*` | 2017-02-28 | `[0%===================]` |

P.S. tasks with `*` will not be considered in release plans.

1. We will freeze feature and prepare to test and release the first version when tasks without `*` are all realized.
2. When developing advanced features, we will provide an option to disable them.

### Persistence Layer (ORM Models)

__Servers__:

Hostname (Primary) | PublicIP | SlavePort | Bandwidth | Transfer | Provider | PrivateIP (Omitempty) | Extra (in JSON)
:-: | :-: | :-: | :-: | :-: | :-: | :-: | :-:

__Services__ (no primary key):

Hostname (Foreign) | Port | Traffic | CreateTime | Status | UserId (Foreign)
:-: | :-: | :-: | :-: | :-: | :-:

__Users__:

UserId (Primary) | Alias | Phone | Email | Password (with salt and md5 hashed) 
:-: | :-: | :-: | :-: | :-:

__Products__:

ProductId (Primary) | Price | Description | Status | Extra (in JSON) 
:-: | :-: | :-: | :-: | :-:

__Orders__:

OrderId (Primary) | Channel | UserId (Foreign) | CreateTime | Amount | ProductId (Foreign)
:-: | :-: | :-: | :-: | :-: | :-:


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


