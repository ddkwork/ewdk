project(user_helloworld)

include(D:/ewdk/dist/sdk/sdk.cmake)

#todo bug No SOURCES given to target: user_helloworld_nav
sdk_add_exe(user_helloworld
        "*.cpp"
        "*.c"
        "*.h"
        "*.asm"
)

#add_clean_target(user_helloworld)#todo bug  another target with the same name already exists.  The existing target is a




