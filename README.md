# go-cryptonote-pool

High performance CryptoNote mining stratum written in Golang.

**Stratum feature list:**

* Concurrent shares processing, each connection is handled in a lightweight thread of execution
* AES-NI enabled share validation code with fallback to slow implementation provided by linking with [**Monero**](https://github.com/monero-project/bitmonero) libraries
* Integrated NewRelic performance monitoring plugin

### Installation

Dependencies:

  * go-1.6
  * Everything required to build bitmonero

Install required packages:

    go get github.com/yvasiyarov/gorelic

#### Mac OS X

Download and compile [Monero](https://github.com/monero-project/bitmonero) daemon.

Now clone stratum repo and compile it:

    git clone https://github.com/sammy007/go-cryptonote-pool.git
    cmake .
    make

Notice that for share validation stratum requires bitmonero source tree where .a libs already compiled. By default stratum will use <code>../bitmonero</code> directory. You can override this behaviour by passing <code>MONERO_DIR</code> env variable:

    MONERO_DIR=/path/to/bitmonero cmake .
    make

#### Linux

I would recommend you to use Ubuntu 16.04 LTS.

In order to successfully link with bitmonero libs, recompile bitmonero with:

    CXXFLAGS="-fPIC" CFLAGS="-fPIC" make release

Build CGO extensions:

    MONERO_DIR=/opt/src/bitmonero cmake .
    make

Build stratum:

    GOPATH=/path/to/go go build -o pool main.go

More info on *GOPATH* you can find in a [wiki](https://github.com/golang/go/wiki/GOPATH).

### Configuration

Configuration is self-describing, just copy *config.example.json* to *config.json* and run stratum with path to config file as 1st argument. There is default XMR address of monero core team in config example and open monero rpc node from [moneroclub.com](https://www.moneroclub.com/node).

### Private Pool Guidelines

For personal private pool you can use [DigitalOcean](https://www.digitalocean.com/?refcode=2a6767e6285f) droplet. With recent blockchain-db merged into Monero it's ok to run it even on 5 USD plan. You will receive 10 USD free credit there.

### TODO

Still in early stage, despite that I am using it for private setups, stratum requires a lot of stability tests. Please run it with <code>-race</code> flag with <code>GORACE="log_path=/path/to/race.log"</code> in private setup and send contents of this file to me if you are "lucky" and found race. It will make stratum ~20x slower, but it does not hit performance if you are soloing with a dozen of GPUs. Look at *-debug.fish* script for example.

Cool stuff will be added after excessive testing, I always have ideas for improvement and new features.

### Donations

* **BTC**: [16bBz4wZPh7kV53nFMf8LmtJHE2rHsADB2](https://blockchain.info/address/16bBz4wZPh7kV53nFMf8LmtJHE2rHsADB2)
* **XMR**: 4Aag5kkRHmCFHM5aRUtfB2RF3c5NDmk5CVbGdg6fefszEhhFdXhnjiTCr81YxQ9bsi73CSHT3ZN3p82qyakHwZ2GHYqeaUr
* **XMR openalias**: wallet.hashinvest.net

### License

Released under the GNU General Public License v2.

http://www.gnu.org/licenses/gpl-2.0.html
