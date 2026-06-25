cmake_minimum_required(VERSION 3.16)

include("C:/Program Files/CMake/bin/ewdk.cmake")

project(KernelDriverDemo)

km_sys(KernelDriverDemo main.c)