# aeon-stratum

High performance Aeon mining stratum with Web-interface written in Golang.

[![Go Report Card](https://goreportcard.com/badge/github.com/sammy007/monero-stratum)](https://goreportcard.com/report/github.com/sammy007/monero-stratum)
[![CircleCI](https://circleci.com/gh/sammy007/monero-stratum/tree/aeon.svg?style=svg)](https://circleci.com/gh/sammy007/monero-stratum/tree/aeon)

**Stratum feature list:**

* Be your own pool
* Rigs availability monitoring
* Keep track of accepts, rejects, blocks stats
* Easy detection of sick rigs
* Daemon failover list
* Concurrent shares processing
* Beautiful Web-interface

![](https://cdn.pbrd.co/images/jRU3qJj83.png)

## Installation

Dependencies:

  * go-1.6
  * Everything required to build Aeon
  * Aeon >= **v0.9.12.0**

### Linux

Use Ubuntu 16.04 LTS.

Compile Aeon source (with a special option):

    git clone https://github.com/aeonix/aeon.git
    cd aeon
    CXXFLAGS="-fPIC" CFLAGS="-fPIC" make

Install Golang and required packages:

    sudo apt-get install golang

Clone stratum:

    git clone https://github.com/sammy007/monero-stratum.git
    cd monero-stratum
    git checkout origin/aeon -b aeon

Build stratum:

    AEON_DIR=/path/to/aeon cmake .
    make

`AEON_DIR=/path/to/monero` is optional, not needed if both `aeon` and `monero-stratum` are in the same parent directory.

### Running Stratum

    ./build/bin/monero-stratum config.json

If you need to bind to privileged ports and don't want to run from `root`:

    sudo apt-get install libcap2-bin
    sudo setcap 'cap_net_bind_service=+ep' /path/to/monero-stratum

## Configuration

Configuration is self-describing, just copy *config.example.json* to *config.json* and run stratum with path to config file as 1st argument.

```javascript
{
  // Address for block rewards
  "address": "WmsydZEeWrSCWCftoSNZF42EzCbkYhSgSZmGVC5QdJB78DFJEQ1XZTcWchV1dVnpHd2m7TjoGHwb5627nc2yVqEe1JSj4Q9q4",
  // Don't validate address
  "bypassAddressValidation": true,
  // Don't validate shares
  "bypassShareValidation": false,

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

  // Aeon daemon
  "upstream": [
    {
      "name": "Main",
      "host": "127.0.0.1",
      "port": 11181,
      "timeout": "10s"
    }
  ]
}
```

### Donations

Donations are welcome.

**XMR**: `4Aag5kkRHmCFHM5aRUtfB2RF3c5NDmk5CVbGdg6fefszEhhFdXhnjiTCr81YxQ9bsi73CSHT3ZN3p82qyakHwZ2GHYqeaUr`
**AEON**: `WmsydZEeWrSCWCftoSNZF42EzCbkYhSgSZmGVC5QdJB78DFJEQ1XZTcWchV1dVnpHd2m7TjoGHwb5627nc2yVqEe1JSj4Q9q4`

### License

Released under the GNU General Public License v2.

http://www.gnu.org/licenses/gpl-2.0.html
