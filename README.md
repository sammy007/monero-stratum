# go-cryptonote-pool

High performance CryptoNote mining stratum with Web-interface written in Golang.

**Stratum feature list:**

* Be your own pool
* Rigs availability monitoring
* Keep track of accepts, rejects, blocks stats
* Easy detection of sick rigs
* Daemon failover list
* Concurrent shares processing
* Beautiful Web-interface

## Installation

Dependencies:

  * go-1.6
  * Everything required to build Monero
  * Monero >= **v0.10.0**

### Linux

Use Ubuntu 16.04 LTS.

Compile Monero source (with libraries option):

    cmake -DBUILD_SHARED_LIBS=1 .
    make

Install Golang and packages:

    sudo apt-get install golang
    export GOPATH=~/go
    go get github.com/yvasiyarov/gorelic

Build CGO extensions:

    MONERO_DIR=/opt/src/monero cmake .
    make

Build stratum:

    go build -o pool main.go

### Mac OS X

Install Golang and packages packages:

    brew update && brew install go
    export GOPATH=~/go
    go get github.com/yvasiyarov/gorelic

Compile Monero source:

    cmake .
    make

Now clone stratum repo and compile it:

    MONERO_DIR=/opt/src/monero cmake .
    make

Build stratum:

    go build -o pool main.go

### Running Stratum

    ./pool config.json

If you need to bind to privileged ports and don't want to run from `root`:

    sudo apt-get install libcap2-bin
    sudo setcap 'cap_net_bind_service=+ep' pool

## Configuration

Configuration is self-describing, just copy *config.example.json* to *config.json* and run stratum with path to config file as 1st argument.

```javascript
{
  // Address for block rewards
  "address": "46BeWrHpwXmHDpDEUmZBWZfoQpdc6HaERCNmx1pEYL2rAcuwufPN9rXHHtyUA4QVy66qeFQkn6sfK8aHYjA3jk3o1Bv16em",
  // Don't validate address
  "bypassAddressValidation": true,
  // Don't validate shares
  "bypassShareValidation": true,

  "threads": 2,

  "estimationWindow": "15m",
  "luckWindow": "24h",
  "largeLuckWindow": "72h",

  // Interval to poll daemon for new jobs
  "blockRefreshInterval": "1s",

  "stratum": {
    // Socket timeout
    "timeout": "15m",

    "listen": [
      {
        "host": "0.0.0.0",
        "port": 1111,
        "diff": 5000,
        "maxConn": 32768
      },
      {
        "host": "0.0.0.0",
        "port": 3333,
        "diff": 10000,
        "maxConn": 32768
      }
    ]
  },

  "frontend": {
    "enabled": true,
    "listen": "0.0.0.0:8082",
    "login": "admin",
    "password": "",
    "hideIP": false
  },

  "upstreamCheckInterval": "5s",

  "upstream": [
    {
      "name": "Main",
      "host": "127.0.0.1",
      "port": 18081,
      "timeout": "10s"
    }
  ]
}
```

### Private Pool Guidelines

For personal private pool you can use [DigitalOcean](https://www.digitalocean.com/?refcode=2a6767e6285f) droplet. With recent blockchain-db merged into Monero it's ok to run it even on 5 USD plan. You will receive 10 USD free credit there.

### Donations

* **BTC**: [16bBz4wZPh7kV53nFMf8LmtJHE2rHsADB2](https://blockchain.info/address/16bBz4wZPh7kV53nFMf8LmtJHE2rHsADB2)
* **XMR**: 4Aag5kkRHmCFHM5aRUtfB2RF3c5NDmk5CVbGdg6fefszEhhFdXhnjiTCr81YxQ9bsi73CSHT3ZN3p82qyakHwZ2GHYqeaUr

### License

Released under the GNU General Public License v2.

http://www.gnu.org/licenses/gpl-2.0.html
