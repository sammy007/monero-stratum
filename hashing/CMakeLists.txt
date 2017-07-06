set(LIB "hashing")

set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -std=c11 -D_GNU_SOURCE")

include_directories("${AEON_DIR}/contrib/epee/include")
include_directories("${AEON_DIR}/src")

link_directories(${AEON_DIR}/build/release/src)

add_library(${LIB} SHARED src/hashing.c)

target_link_libraries(${LIB} libcrypto.a)
