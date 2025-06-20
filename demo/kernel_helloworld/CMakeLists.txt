cmake_minimum_required(VERSION 3.20)
project(kernel_helloworld)

include(D:/ewdk/dist/wdk/wdk.cmake)

wdk_add_exe(kernel_helloworld
        "*.c"
        "*.h"
        "*.asm"
)

wdk_add_clean_all_target()




