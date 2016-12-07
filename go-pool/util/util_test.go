package util

import (
	"encoding/hex"
	"math/big"
	"testing"
)

func TestGetTargetHex(t *testing.T) {
	targetHex := GetTargetHex(500)
	expectedHex := "6e128300"
	if targetHex != expectedHex {
		t.Error("Invalid targetHex")
	}

	targetHex = GetTargetHex(15000)
	expectedHex = "7b5e0400"
	if targetHex != expectedHex {
		t.Error("Invalid targetHex")
	}
}

func TestGetHashDifficulty(t *testing.T) {
	hash := "8e3c1865f22801dc3df0a688da80701e2390e7838e65c142604cc00eafe34000"
	hashBytes, _ := hex.DecodeString(hash)
	diff := new(big.Int)
	diff.SetBytes(reverse(hashBytes))
	shareDiff := GetHashDifficulty(hashBytes)

	if shareDiff.String() != "1009" {
		t.Error("Invalid diff")
	}
}
