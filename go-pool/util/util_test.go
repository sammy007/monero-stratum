package util

import (
	"encoding/hex"
	"math/big"
	"testing"
)

func TestGetTargetHex(t *testing.T) {
	target, targetHex := GetTargetHex(500)
	expectedTarget := uint32(1846706944)
	expectedHex := "6e128300"
	if target != expectedTarget {
		t.Error("Invalid target")
	}
	if targetHex != expectedHex {
		t.Error("Invalid targetHex")
	}

	target, targetHex = GetTargetHex(15000)
	expectedTarget = uint32(2069758976)
	expectedHex = "7b5e0400"
	if target != expectedTarget {
		t.Error("Invalid target")
	}
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
