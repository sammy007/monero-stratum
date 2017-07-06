set(CXXLIB "cnutil")

find_package(Boost COMPONENTS thread system program_options date_time filesystem REQUIRED)

# Flags
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -std=c++11 -D_GNU_SOURCE")

include_directories(${Boost_INCLUDE_DIRS})
include_directories("${AEON_DIR}/contrib/epee/include")
include_directories("${AEON_DIR}/src")

link_directories(${AEON_DIR}/build/release/src)

# Build library
add_library(${CXXLIB} SHARED src/cnutil.cpp)

target_link_libraries(${CXXLIB}
    libcryptonote_core.a
    libcrypto.a
    libcommon.a
    libringct.a
)

target_link_libraries(${CXXLIB}
    ${Boost_THREAD_LIBRARY}
    ${Boost_SYSTEM_LIBRARY}
    ${Boost_PROGRAM_OPTIONS_LIBRARY}
    ${Boost_DATE_TIME_LIBRARY}
    ${Boost_FILESYSTEM_LIBRARY}
)
