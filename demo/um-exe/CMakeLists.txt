cmake_minimum_required(VERSION 3.16)

include("C:/Program Files/CMake/bin/ewdk.cmake")

project(UmExeDemo)

um_exe(UmExeDemo SUBSYSTEM WINCON main.c)