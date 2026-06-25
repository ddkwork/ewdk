cmake_minimum_required(VERSION 3.28)

include("C:/Program Files/CMake/bin/ewdk.cmake")

project(MfcDemo)

set(CMAKE_CXX_STANDARD 17)

um_exe_mfc(MfcDemo
    SOURCES main.cpp
)
