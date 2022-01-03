# Proxy
Minimalistic TCP relay proxy.

## Installation
- ensure you have go >= 1.17 installed
- clone the repo
- `cd proxy`
- `go install main.go`

## Examples
Listen on port 7777 on **all interfces** and relay to port 3333 on localhost
```
proxy :7777 :3333
```


Listen on port 7777 on **localhost** and relay to port 3333 on localhost
```
proxy localhost:7777 :3333
```


Listen on port 7777 on all interfaces and relay to port 80 on example.com using a 30 seconds timeout.
The following are all equivalent
```
proxy -timeout=30 :7777 example.com:80
proxy -timeout=30 :7777 example.com:http
proxy -timeout=30 0.0.0.0:7777 example.com:80
proxy -timeout=30 0.0.0.0:7777 example.com:http
```


## Usage
```
Usage: proxy [OPTIONS] [bind_address]:bind_port [target_host]:target_port

Description:
Proxy relays TCP connections. For every incoming
connection, a new connection is established to the target
host and the data is relayed between the two connections.

OPTIONS:
  -show-active
        When true, prints the number of active connections. (default true)
  -timeout uint
        Timeout in seconds for dialing to target host. (default 10)
```