set(CXXLIB "cnutilxx")

find_package(Boost COMPONENTS thread system program_options date_time filesystem REQUIRED)

# Flags
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -std=c++11 -D_GNU_SOURCE")

include_directories(${Boost_INCLUDE_DIRS})
include_directories("${MONERO_DIR}/contrib/epee/include")
include_directories("${MONERO_DIR}/src")

# Build library
add_library(${CXXLIB} SHARED cnutilxx/main.cpp)

target_link_libraries(${CXXLIB}
    ${MONERO_DIR}/build/release/src/cryptonote_core/libcryptonote_core.a
    ${MONERO_DIR}/build/release/src/crypto/libcrypto.a
    ${MONERO_DIR}/build/release/src/common/libcommon.a
)

target_link_libraries(${CXXLIB}
    ${Boost_THREAD_LIBRARY}
    ${Boost_SYSTEM_LIBRARY}
    ${Boost_PROGRAM_OPTIONS_LIBRARY}
    ${Boost_DATE_TIME_LIBRARY}
    ${Boost_FILESYSTEM_LIBRARY}
)

set(LIB "cnutil")

# Flags
set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -std=c11 -D_GNU_SOURCE")

# Build library
add_library(${LIB} SHARED cnutil.c)
target_link_libraries(${LIB} ${CXXLIB})
