cmake_minimum_required(VERSION 3.20)
project(MyAdvancedDriver)

include(wdk64/wdk.cmake)

wdk_add_driver(MyAdvancedDriver
        "*.c"
        "*.h"
        "*.asm"
)

wdk_add_clean_all_target()




