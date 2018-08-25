package util

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/big"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sammy007/monero-stratum/cnutil"
)

var Diff1 = StringToBig("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")

func StringToBig(h string) *big.Int {
	n := new(big.Int)
	n.SetString(h, 0)
	return n
}

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func GetTargetHex(diff int64) string {
	padded := make([]byte, 32)

	diffBuff := new(big.Int).Div(Diff1, big.NewInt(diff)).Bytes()
	copy(padded[32-len(diffBuff):], diffBuff)
	buff := padded[0:4]
	targetHex := hex.EncodeToString(reverse(buff))
	return targetHex
}

func GetHashDifficulty(hashBytes []byte) (*big.Int, bool) {
	diff := new(big.Int)
	diff.SetBytes(reverse(hashBytes))

	// Check for broken result, empty string or zero hex value
	if diff.Cmp(new(big.Int)) == 0 {
		return nil, false
	}
	return diff.Div(Diff1, diff), true
}

func ValidateAddress(addy string, poolAddy string) bool {
	if len(addy) != len(poolAddy) {
		return false
	}
	prefix, _ := utf8.DecodeRuneInString(addy)
	poolPrefix, _ := utf8.DecodeRuneInString(poolAddy)
	if prefix != poolPrefix {
		return false
	}
	return cnutil.ValidateAddress(addy)
}

func reverse(src []byte) []byte {
	dst := make([]byte, len(src))
	for i := len(src); i > 0; i-- {
		dst[len(src)-i] = src[i-1]
	}
	return dst
}

// load any JSON file into a struct and return it
func LoadJson(filename string) interface{} {
	var s interface{}
	data, err := ioutil.ReadFile(filename)
	if err == nil {
		// yuk.. what am I missing here...?
		decoder := json.NewDecoder(strings.NewReader(string(data)))

		// we have to set UseNumber to avoid getting float64s for all our timestamps, etc
		decoder.UseNumber()
		err := decoder.Decode(&s)

		if err != nil {
			log.Println("parsing stats file", err.Error())
		}

	} else {
		log.Println("opening json file", err.Error())
	}

	return s

}

// save any Struct to a JSON file
func SaveJson(filename string, s interface{}) {
	jsonData, err := json.Marshal(s)
	if err == nil {
		err = ioutil.WriteFile(filename, []byte(jsonData), 0644)
		if err != nil {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}

}