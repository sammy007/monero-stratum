#include "crypto/hash-ops.h"

void cryptonight_hash(const char* input, char* output, uint32_t len) {
    cn_slow_hash(input, len, output);
}

void cryptonight_fast_hash(const char* input, char* output, uint32_t len) {
    cn_fast_hash(input, len, output);
}
