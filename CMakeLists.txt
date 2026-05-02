cmake_minimum_required(VERSION 3.22)

include("C:/Program Files/CMake/bin/ewdk.cmake")

project(AllDemos)
set(CMAKE_CXX_STANDARD 14)

add_subdirectory(demo/kernel)
add_subdirectory(demo/um-lib)
add_subdirectory(demo/um-dll)
add_subdirectory(demo/um-exe)
add_subdirectory(demo/NTMemory)
