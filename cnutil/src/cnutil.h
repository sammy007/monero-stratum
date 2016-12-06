#include <stdint.h>
#include "stdbool.h"

#ifdef __cplusplus
extern "C" {
#endif

uint32_t convert_blob(const char *blob, uint32_t len, char *out);
bool validate_address(const char *addr, uint32_t len);

#ifdef __cplusplus
}
#endif
