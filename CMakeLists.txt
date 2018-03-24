cmake_minimum_required (VERSION 2.8.11)
project(pool)

if (DEFINED ENV{AEON_DIR})
    get_filename_component(AEON_DIR $ENV{AEON_DIR} ABSOLUTE)
    message("Using Monero source from env ${AEON_DIR}")
else()
    get_filename_component(AEON_DIR "${CMAKE_SOURCE_DIR}/../monero" ABSOLUTE)
    message("Monero surce directory is not defined, using default ${AEON_DIR}")
endif()

add_subdirectory(hashing)
add_subdirectory(cnutil)

add_custom_command(
    OUTPUT build/bin
    COMMAND build/env.sh go get -v ./...
)

add_custom_target(build ALL DEPENDS hashing cnutil ${CMAKE_CURRENT_BINARY_DIR}/build/bin)
