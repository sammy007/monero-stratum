#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>
#include "stdbool.h"

uint32_t convert_blob(const char *blob, uint32_t len, char *out);
bool validate_address(const char *addr, uint32_t len);
