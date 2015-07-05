#include <stdint.h>
#include <string>
#include "cryptonote_core/cryptonote_format_utils.h"
#include "common/base58.h"

using namespace cryptonote;

// Well, it's dirty and useless, but without it I can't link to bitmonero's /build/release/src/**/*.a libs
unsigned int epee::g_test_dbg_lock_sleep = 0;

extern "C" uint32_t cn_convert_blob(const char *blob, size_t len, char *out) {
    std::string input = std::string(blob, len);
    std::string output = "";

    block b = AUTO_VAL_INIT(b);
    if (!parse_and_validate_block_from_blob(input, b)) {
        return 0;
    }

    output = get_block_hashing_blob(b);
    output.copy(out, output.length(), 0);
    return output.length();
}

extern "C" bool cn_validate_address(const char *addr, size_t len) {
    std::string input = std::string(addr, len);
    std::string output = "";
    uint64_t prefix;
    return tools::base58::decode_addr(addr, prefix, output);
}
