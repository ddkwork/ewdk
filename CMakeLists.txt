cmake_minimum_required(VERSION 3.22)
project(AllDemos)
set(CMAKE_CXX_STANDARD 14)
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS}")
set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS}")

list(APPEND CMAKE_MODULE_PATH "${CMAKE_CURRENT_SOURCE_DIR}")
include(D:/ewdk/ewdk.cmake)

add_subdirectory(demo/kernel)
add_subdirectory(demo/um-lib)
add_subdirectory(demo/um-dll)
add_subdirectory(demo/um-exe)