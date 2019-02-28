package hashing

import "testing"
import "log"
import "encoding/hex"

func TestHash(t *testing.T) {
	blob, _ := hex.DecodeString("a1009091e4aa05ff5fe4801727ed0c1b8b339e1a0054d75568fec6ba9c4346e88b10d59edbf6858b2b00008a63b2865b65b84d28bb31feb057b16a21e2eda4bf6cc6377e3310af04debe4a01")
	hashBytes := Hash(blob, false, 0)
	hash := hex.EncodeToString(hashBytes)
	log.Println(hash)

	expectedHash := "9d38ea4c527970fee1c4f89272a6fb787de71fc95bba7592eaed6e20b7f7478e"

	if hash != expectedHash {
		t.Error("Invalid hash")
	}
}

func TestHash_fast(t *testing.T) {
	blob, _ := hex.DecodeString("01009091e4aa05ff5fe4801727ed0c1b8b339e1a0054d75568fec6ba9c4346e88b10d59edbf6858b2b00008a63b2865b65b84d28bb31feb057b16a21e2eda4bf6cc6377e3310af04debe4a01")
	hashBytes := Hash(blob, true, 0)
	hash := hex.EncodeToString(hashBytes)
	log.Println(hash)

	expectedHash := "7591f4d8ff9d86ea44873e89a5fb6f380f4410be6206030010567ac9d0d4b0e1"

	if hash != expectedHash {
		t.Error("Invalid fast hash")
	}
}

func TestFastHash(t *testing.T) {
	blob, _ := hex.DecodeString("01009091e4aa05ff5fe4801727ed0c1b8b339e1a0054d75568fec6ba9c4346e88b10d59edbf6858b2b00008a63b2865b65b84d28bb31feb057b16a21e2eda4bf6cc6377e3310af04debe4a01")
	hashBytes := FastHash(blob)
	hash := hex.EncodeToString(hashBytes)
	log.Println(hash)

	expectedHash := "8706c697d9fc8a48b14ea93a31c6f0750c48683e585ec1a534e9c57c97193fa6"

	if hash != expectedHash {
		t.Error("Invalid fast hash")
	}
}

func BenchmarkHash(b *testing.B) {
	blob, _ := hex.DecodeString("a100b69bb3aa050a3106491f858f8646d3a8d13fd9924403bf07af95e6e7cc9e4ad105d76da27241565555866b1baa9db8f027cf57cd45d6835c11287b210d9ddb407deda565f8112e19e501")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Hash(blob, false, 0)
	}
}

func BenchmarkHashParallel(b *testing.B) {
	blob, _ := hex.DecodeString("a100b69bb3aa050a3106491f858f8646d3a8d13fd9924403bf07af95e6e7cc9e4ad105d76da27241565555866b1baa9db8f027cf57cd45d6835c11287b210d9ddb407deda565f8112e19e501")
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Hash(blob, false, 0)
		}
	})
}

func BenchmarkHash_fast(b *testing.B) {
	blob, _ := hex.DecodeString("0100b69bb3aa050a3106491f858f8646d3a8d13fd9924403bf07af95e6e7cc9e4ad105d76da27241565555866b1baa9db8f027cf57cd45d6835c11287b210d9ddb407deda565f8112e19e501")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Hash(blob, true, 0)
	}
}

func BenchmarkFastHash(b *testing.B) {
	blob, _ := hex.DecodeString("0100b69bb3aa050a3106491f858f8646d3a8d13fd9924403bf07af95e6e7cc9e4ad105d76da27241565555866b1baa9db8f027cf57cd45d6835c11287b210d9ddb407deda565f8112e19e501")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		FastHash(blob)
	}
}
