#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

uint32_t cn_convert_blob(const char *blob, uint32_t len, char *out);
bool cn_validate_address(const char *addr, uint32_t len);

#ifdef __cplusplus
}
#endif
