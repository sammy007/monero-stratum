package util

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"math/rand"
	"strconv"
	"time"
	"unicode/utf8"

	"../../cnutil"
)

var Diff1 *big.Int

func init() {
	Diff1 = new(big.Int)
	Diff1.SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", 16)
}

func Random() string {
	min := int64(100000000000000)
	max := int64(999999999999999)
	n := rand.Int63n(max-min+1) + min
	return strconv.FormatInt(n, 10)
}

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func GetTargetHex(diff int64) (uint32, string) {
	padded := make([]byte, 32)

	diff2 := new(big.Int)
	diff2.SetInt64(int64(diff))

	diff3 := new(big.Int)
	diff3 = diff3.Div(Diff1, diff2)

	diffBuff := diff3.Bytes()
	copy(padded[32-len(diffBuff):], diffBuff)
	buff := padded[0:4]
	var target uint32
	targetBuff := bytes.NewReader(buff)
	binary.Read(targetBuff, binary.LittleEndian, &target)
	targetHex := hex.EncodeToString(reverse(buff))

	return target, targetHex
}

func GetHashDifficulty(hashBytes []byte) *big.Int {
	diff := new(big.Int)
	diff.SetBytes(reverse(hashBytes))
	return diff.Div(Diff1, diff)
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

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func reverse(src []byte) []byte {
	dst := make([]byte, len(src))
	for i := len(src); i > 0; i-- {
		dst[len(src)-i] = src[i-1]
	}
	return dst
}
