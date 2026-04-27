cmake_minimum_required(VERSION 3.22)

if(EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/ewdk-env.cmake")
    include(${CMAKE_CURRENT_SOURCE_DIR}/ewdk-env.cmake)
endif()

list(APPEND CMAKE_MODULE_PATH "${CMAKE_CURRENT_SOURCE_DIR}")
include(${CMAKE_CURRENT_SOURCE_DIR}/ewdk.cmake)

project(AllDemos)
set(CMAKE_CXX_STANDARD 14)

add_subdirectory(demo/kernel)
add_subdirectory(demo/um-lib)
add_subdirectory(demo/um-dll)
add_subdirectory(demo/um-exe)
