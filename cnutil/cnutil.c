#include <stdint.h>
#include "stdbool.h"
#include <stdio.h>
#include <stdlib.h>
#include "cnutilxx/main.h"

uint32_t convert_blob(const char *blob, size_t len, char *out) {
    return cn_convert_blob(blob, len, out);
}

bool validate_address(const char *addr, size_t len) {
    return cn_validate_address(addr, len);
}
