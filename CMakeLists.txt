cmake_minimum_required (VERSION 2.8.11)
project (pool)

if (DEFINED ENV{AEON_DIR})
    get_filename_component(AEON_DIR $ENV{AEON_DIR} ABSOLUTE)
    message("Using Aeon source from env ${AEON_DIR}")
else()
    get_filename_component(AEON_DIR "${CMAKE_SOURCE_DIR}/../aeon" ABSOLUTE)
    message("Aeon source directory is not defined, using default ${AEON_DIR}")
endif()

add_subdirectory(hashing)
add_subdirectory(cnutil)
