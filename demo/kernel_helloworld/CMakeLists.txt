project(kernel_helloworld)

include(D:/ewdk/dist/wdk/wdk.cmake)

wdk_add_driver(kernel_helloworld
        "*.c"
        "*.h"
        "*.asm"
)

wdk_add_clean_all_target()




