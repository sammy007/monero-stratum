set(CXXLIB "cnutil")

find_package(Boost COMPONENTS thread system program_options date_time filesystem REQUIRED)

# Flags
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -std=c++11 -D_GNU_SOURCE")

include_directories(${Boost_INCLUDE_DIRS})
include_directories("${MONERO_DIR}/contrib/epee/include")
include_directories("${MONERO_DIR}/src")
include_directories("${MONERO_DIR}/external/easylogging++")

link_directories(
    ${MONERO_DIR}/build/release/src/cryptonote_core
    ${MONERO_DIR}/build/release/src/cryptonote_basic
    ${MONERO_DIR}/build/release/src/crypto
    ${MONERO_DIR}/build/release/src/common
    ${MONERO_DIR}/build/release/src/ringct
    ${MONERO_DIR}/build/release/contrib/epee/src
    ${MONERO_DIR}/build/release/external/easylogging++
)

# Build library
add_library(${CXXLIB} SHARED src/cnutil.cpp)

target_link_libraries(${CXXLIB}
    cryptonote_core
    cryptonote_basic
    cncrypto
    common
    ringct
    epee
    easylogging
)

target_link_libraries(${CXXLIB}
    ${Boost_THREAD_LIBRARY}
    ${Boost_SYSTEM_LIBRARY}
    ${Boost_PROGRAM_OPTIONS_LIBRARY}
    ${Boost_DATE_TIME_LIBRARY}
    ${Boost_FILESYSTEM_LIBRARY}
)
