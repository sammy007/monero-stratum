# go-cryptonote-pool

High performance CryptoNote mining stratum written in Golang backed by Redis.

**Stratum feature list:**

* Full [node-cryptonote-pool](https://github.com/zone117x/node-cryptonote-pool) database compatibility
* Concurrent shares processing, each connection is handled in a lightweight thread of execution
* Several configurable stratum policies to prevent basic attacks
* Banning policy using [**ipset**s](http://ipset.netfilter.org/) on Linux for high performance banning
* Whitelist for trusted miners and blacklist for unwelcome guests
* AES-NI enabled share validation code with fallback to slow implementation provided by linking with [**Monero**](https://github.com/monero-project/bitmonero) libraries
* Integrated NewRelic performance monitoring plugin

### Installation

Dependencies:

  * go-1.4
  * boost-1.55+
  * cmake

Install required packages:

    go get gopkg.in/redis.v3
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

Installation on linux is similar to OS X installation and currently the only dfference is that you should copy *.so* libs from *hashing* and *cnutil* directories to */usr/local/lib* or similar dir in order to make CGO happy. I would recommend you to use Ubuntu 14.04 LTS.

In order to successfully link with bitmonero libs, recompile bitmonero with:

    CXXFLAGS="-fPIC" CFLAGS="-fPIC" make release

Build stratum:

    MONERO_DIR=/opt/src/bitmonero cmake .
    make

Run it:

    LD_LIBRARY_PATH="/usr/local/lib/" GOPATH=/path/to/go go run main.go

More info on *GOPATH* you can find in a [wiki](https://github.com/golang/go/wiki/GOPATH).

### Configuration

Configuration is self-describing, just copy *config.example.json* to *config.json* and run stratum with path to config file as 1st argument. There is default XMR address of monero core team in config example and open monero rpc node from [moneroclub.com](https://www.moneroclub.com/node).

#### Redis

Leave Redis password blank if you have local setup in a trusted environment. Don't rely on Redis password, it's easily bruteforceable. Password option is only for some clouds. There is a connection pool, use some reasonable value. Remember, that each valid share submission will lease one connection from a pool due to <code>multi</code> exec and instantly release it, this is how go-redis works.

#### Policies

Stratum policy server collecting several stats on per IP basis.

Banning enabled by default. Specify <code>ipset</code> name for banning. Timeout argument will be passed to this ipset. For ipset usage refer to [this article](https://wiki.archlinux.org/index.php/Ipset). Stratum will use os/exec command like <code>sudo ipset ...</code> for banning, so you have to configure sudo properly and make sure that your system will never ask for password:

*/etc/sudoers.d/stratum*

    stratum ALL=NOPASSWD: /sbin/ipset

Use limits to prevent connection flood to your stratum, there is initial <code>limit</code> and <code>limitJump</code>. Policy server will increase number of allowed connections on each valid share submission. Stratum will bypass this policy regarding <code>grace</code> time specified on start.

#### Payouts and Block Unlocking

This is just stratum yet. Use corresponding [node-cryptonote-pool](https://github.com/zone117x/node-cryptonote-pool) modules for block unlocking and payout processing. Database is 100% compatible.

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
