package hashing

// #cgo CFLAGS: -std=c11 -D_GNU_SOURCE
// #cgo LDFLAGS: -L${SRCDIR} -lhashing -Wl,-rpath ${SRCDIR} -lstdc++
// #include <stdlib.h>
// #include <stdint.h>
// #include "src/hashing.h"
import "C"
import "unsafe"

func Hash(blob []byte, fast bool, height int64) []byte {
	output := make([]byte, 32)
	if fast {
		C.cryptonight_fast_hash((*C.char)(unsafe.Pointer(&blob[0])), (*C.char)(unsafe.Pointer(&output[0])), (C.uint32_t)(len(blob)))
	} else {
		C.cryptonight_hash((*C.char)(unsafe.Pointer(&blob[0])), (*C.char)(unsafe.Pointer(&output[0])), (C.uint32_t)(len(blob)), (C.uint64_t)(uint64(height)))
	}
	return output
}

func FastHash(blob []byte) []byte {
	return Hash(append([]byte{byte(len(blob))}, blob...), true, 0)
}
