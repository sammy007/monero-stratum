# go-cryptonote-pool

High performance CryptoNote mining stratum written in Golang.

**Stratum feature list:**

* Concurrent shares processing
* AES-NI enabled share validation code with fallback to slow implementation
* Integrated NewRelic performance monitoring plugin

## Installation

Dependencies:

  * go-1.6
  * Everything required to build monero
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
    // Address to where mined blocks will rain
    "address": "46BeWrHpwXmHDpDEUmZBWZfoQpdc6HaERCNmx1pEYL2rAcuwufPN9rXHHtyUA4QVy66qeFQkn6sfK8aHYjA3jk3o1Bv16em",
    // Don't validate login, useful for other CN coins
    "bypassAddressValidation": false,
    // Don't validate shares for efficiency
    "bypassShareValidation": false,

    "threads": 2,

    // Mining endpoints
    "stratum": {
        // TCP timeout for miner, better keep default
        "timeout": "15m",
        // Interval to poll monero node for new jobs
        "blockRefreshInterval": "1s",

        "listen": [
            {
                "host": "0.0.0.0",
                "port": 1111,
                // Stratum port static difficulty
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

    // Monero daemon connection options
    "daemon": {
        "host": "127.0.0.1",
        // Monero RPC port, default is 18081
        "port": 18081,
        "timeout": "10s"
    }
}
```

### Private Pool Guidelines

For personal private pool you can use [DigitalOcean](https://www.digitalocean.com/?refcode=2a6767e6285f) droplet. With recent blockchain-db merged into Monero it's ok to run it even on 5 USD plan. You will receive 10 USD free credit there.

### TODO

In-RAM stats with a simple self hosted frontend.

### Donations

* **BTC**: [16bBz4wZPh7kV53nFMf8LmtJHE2rHsADB2](https://blockchain.info/address/16bBz4wZPh7kV53nFMf8LmtJHE2rHsADB2)
* **XMR**: 4Aag5kkRHmCFHM5aRUtfB2RF3c5NDmk5CVbGdg6fefszEhhFdXhnjiTCr81YxQ9bsi73CSHT3ZN3p82qyakHwZ2GHYqeaUr

### License

Released under the GNU General Public License v2.

http://www.gnu.org/licenses/gpl-2.0.html
