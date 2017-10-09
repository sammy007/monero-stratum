package cnutil

import (
	"encoding/hex"
	"testing"
)

func TestConvertBlob(t *testing.T) {
	hashBytes, _ := hex.DecodeString("0100a5d1fca9057dff46d140d453a672437ba0ec7d6a74bc5fa0391f8a918e41fd7ba2cf6fc1af000000000183811401ffc7801405889ec5dc2402e0fe0db63a8e532a7b988e0c32a764e5e8d64d7efac9bc9d24ce32b0984ab93980b09dc2df0102ccd38432501a9182ccc5b44cb47abdddbc9a6321cd581f5f07a7a9195795d5c68080dd9da41702f6e944ee4c6e1eeaed1fa6a3c2480a410e959c6e823b96dad54d31fe223cc0fd80c0a8ca9a3a0286d0e3411670e4c2abe8492c695c66d8262660ee88a0b14a2b03c9180fb6f0d480c0caf384a30202aec5c9b7efe841dd821476e0e06217be13a4c85a83efcf9576314d60130e02e72b0150526f7a381cec33e5827c1848dd80e6eac4b262304ea06b3a43303a4631df28020800000000018ba82000")
	expectedResult, _ := hex.DecodeString("0100a5d1fca9057dff46d140d453a672437ba0ec7d6a74bc5fa0391f8a918e41fd7ba2cf6fc1af00000000e81cb2bf0d2c5054a49bda094c39cb263a9565b9b81cf4c4f848292040419f4a01")
	output := ConvertBlob(hashBytes)

	if len(output) != 76 {
		t.Error("Invalid result length")
	}
	ok := true
	for i := range output {
		if expectedResult[i] != output[i] {
			ok = false
			break
		}
	}
	if !ok {
		t.Error("Invalid result")
	}
}

func TestDecodeAddress(t *testing.T) {
	addy := "45pyCXYn2UBVUmCFjgKr7LF8hCTeGwucWJ2xni7qrbj6GgAZBFY6tANarozZx9DaQqHyuR1AL8HJbRmqwLhUaDpKJW4hqS1"
	if !ValidateAddress(addy) {
		t.Error("Valid address")
	}

	addy = "46BeWrHpwXmHDpDEUmZBWZfoQpdc6HaERCNmx1pEYL2rAcuwufPN9rXHHtyUA4QVy66qeFQkn6sfK8aHYjA3jk3o1Bv16em"
	if !ValidateAddress(addy) {
		t.Error("Valid address")
	}

	if ValidateAddress("OMG") {
		t.Error("Invalid address")
	}
}

func BenchmarkConvertBlob(b *testing.B) {
	for i := 0; i < b.N; i++ {
		hashBytes, _ := hex.DecodeString("0100a5d1fca9057dff46d140d453a672437ba0ec7d6a74bc5fa0391f8a918e41fd7ba2cf6fc1af000000000183811401ffc7801405889ec5dc2402e0fe0db63a8e532a7b988e0c32a764e5e8d64d7efac9bc9d24ce32b0984ab93980b09dc2df0102ccd38432501a9182ccc5b44cb47abdddbc9a6321cd581f5f07a7a9195795d5c68080dd9da41702f6e944ee4c6e1eeaed1fa6a3c2480a410e959c6e823b96dad54d31fe223cc0fd80c0a8ca9a3a0286d0e3411670e4c2abe8492c695c66d8262660ee88a0b14a2b03c9180fb6f0d480c0caf384a30202aec5c9b7efe841dd821476e0e06217be13a4c85a83efcf9576314d60130e02e72b0150526f7a381cec33e5827c1848dd80e6eac4b262304ea06b3a43303a4631df28020800000000018ba82000")
		ConvertBlob(hashBytes)
	}
}

func BenchmarkDecodeAddress(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateAddress("45pyCXYn2UBVUmCFjgKr7LF8hCTeGwucWJ2xni7qrbj6GgAZBFY6tANarozZx9DaQqHyuR1AL8HJbRmqwLhUaDpKJW4hqS1")
	}
}
