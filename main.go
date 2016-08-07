package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"./go-pool/pool"
	"./go-pool/stratum"

	"github.com/yvasiyarov/gorelic"
)

var cfg pool.Config

func startStratum() {
	if cfg.Threads > 0 {
		runtime.GOMAXPROCS(cfg.Threads)
		log.Printf("Running with %v threads", cfg.Threads)
	} else {
		n := runtime.NumCPU()
		runtime.GOMAXPROCS(n)
		log.Printf("Running with default %v threads", n)
	}

	quit := make(chan bool)
	for _, port := range cfg.Stratum.Ports {
		s := stratum.NewStratum(&cfg, port)

		go func() {
			s.Listen()
			quit <- true
		}()
	}
	<-quit
}

func startNewrelic() {
	// Run NewRelic
	if cfg.NewrelicEnabled {
		nr := gorelic.NewAgent()
		nr.Verbose = cfg.NewrelicVerbose
		nr.NewrelicLicense = cfg.NewrelicKey
		nr.NewrelicName = cfg.NewrelicName
		nr.Run()
	}
}

func readConfig(cfg *pool.Config) {
	configFileName := "config.json"
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}
	configFileName, _ = filepath.Abs(configFileName)
	log.Printf("Loading config: %v", configFileName)

	configFile, err := os.Open(configFileName)
	if err != nil {
		log.Fatal("File error: ", err.Error())
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&cfg); err != nil {
		log.Fatal("Config error: ", err.Error())
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	readConfig(&cfg)
	startNewrelic()
	startStratum()
}
