set(LIB "hashing")

set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -std=c11 -D_GNU_SOURCE")

include_directories("${MONERO_DIR}/contrib/epee/include")
include_directories("${MONERO_DIR}/src")

add_library(${LIB} SHARED src/hashing.c)
target_link_libraries(${LIB}
    ${MONERO_DIR}/build/release/src/crypto/libcrypto.a
)
