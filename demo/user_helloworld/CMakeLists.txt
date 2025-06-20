cmake_minimum_required(VERSION 3.20)
project(user_helloworld)

include(D:/ewdk/dist/wdk/wdk.cmake)

wdk_add_driver(user_helloworld
        "*.cpp"
        "*.c"
        "*.h"
        "*.asm"
)

wdk_add_clean_all_target()




