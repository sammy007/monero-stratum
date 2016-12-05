cmake_minimum_required (VERSION 2.8.11)
project (pool)

if (DEFINED ENV{MONERO_DIR})
    get_filename_component(MONERO_DIR $ENV{MONERO_DIR} ABSOLUTE)
    message("Using Monero source from env ${MONERO_DIR}")
else()
    get_filename_component(MONERO_DIR "${CMAKE_SOURCE_DIR}/../monero" ABSOLUTE)
    message("Monero surce directory is not defined, using default ${MONERO_DIR}")
endif()

add_subdirectory(hashing)
add_subdirectory(cnutil)
