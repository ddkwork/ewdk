cmake_minimum_required(VERSION 3.16)
project(UmExeDemo)

wdk_add_executable(UmExeDemo SUBSYSTEM WINCON main.c)
