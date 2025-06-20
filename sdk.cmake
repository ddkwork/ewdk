# sdk_build.cmake - 完整IDE支持的Windows SDK构建系统

# ====================== 全局配置 ======================
set(SDK_ROOT "D:/ewdk/dist/sdk" CACHE PATH "SDK根目录")
set(SDK_VERSION "10.0.26100.0" CACHE STRING "SDK版本")
set(DEFAULT_EXE_ENTRY "main" CACHE STRING "EXE默认入口点")
set(DEFAULT_DLL_ENTRY "DllMain" CACHE STRING "DLL默认入口点")
set(OUTPUT_DIR "${CMAKE_BINARY_DIR}/bin" CACHE PATH "输出目录")
set(OBJ_DIR "${CMAKE_BINARY_DIR}/obj" CACHE PATH "对象文件目录")

# ====================== 工具链映射 ======================
# 统一工具链配置结构
set(TOOLCHAIN_CONFIG
    # x64工具链
    x64:BIN_HOST "Hostx64"
    x64:CL_EXE "bin/Hostx64/x64/cl.exe"
    x64:LINK_EXE "bin/Hostx64/x64/link.exe"
    x64:LIB_EXE "lib/x64/lib.exe"
    x64:ML_EXE "bin/Hostx64/x64/ml64.exe"
    x64:LIB_PATHS "lib/x64" "dia/lib/amd64"

    # x86工具链
    x86:BIN_HOST "Hostx64"
    x86:CL_EXE "bin/Hostx64/x86/cl.exe"
    x86:LINK_EXE "bin/Hostx64/x86/link.exe"
    x86:LIB_EXE "lib/x86/lib.exe"
    x86:ML_EXE "bin/Hostx64/x86/ml.exe"
    x86:LIB_PATHS "lib/x86"
)

# ====================== 包含目录 ======================
# 共享的包含目录结构
set(SDK_INCLUDE_PATHS
    "${SDK_ROOT}/include"
    "${SDK_ROOT}/dia/include"

    # WDK专用包含目录
    "D:/ewdk/dist/wdk/Include/wdf/kmdf/1.35"
    "D:/ewdk/dist/wdk/Include/10.0.26100.0/shared"
    "D:/ewdk/dist/wdk/Include/10.0.26100.0/km"
    "D:/ewdk/dist/wdk/Include/10.0.26100.0/km/crt"
)

# ====================== 架构支持 ======================
set(SUPPORTED_ARCH x64 x86 CACHE STRING "支持的架构")
set_property(CACHE SUPPORTED_ARCH PROPERTY STRINGS x64 x86)
set(SDK_ARCH "x64" CACHE STRING "目标架构")

# ====================== IDE导航目标 ======================
# 用于CLion等IDE的导航目标
function(add_navigation_target target)
    cmake_parse_arguments(ARG "" "ARCH" "SOURCES" ${ARGN})

    # 创建伪目标，仅用于IDE导航
    add_library(${target}_nav STATIC ${ARG_SOURCES})
    set_target_properties(${target}_nav PROPERTIES
        EXCLUDE_FROM_ALL TRUE
        EXCLUDE_FROM_DEFAULT_BUILD TRUE
        RUNTIME_OUTPUT_DIRECTORY ${OUTPUT_DIR}/nav
        ARCHIVE_OUTPUT_DIRECTORY ${OUTPUT_DIR}/nav
        LIBRARY_OUTPUT_DIRECTORY ${OUTPUT_DIR}/nav
    )

    # 添加包含目录
    target_include_directories(${target}_nav PRIVATE
        ${SDK_INCLUDE_PATHS}
    )

    # 为IDE设置架构宏
    if(ARG_ARCH STREQUAL "x64")
        target_compile_definitions(${target}_nav PRIVATE
            _WIN64 _AMD64_ AMD64
            __x86_64 __x86_64__
        )
    else()
        target_compile_definitions(${target}_nav PRIVATE
            _WIN32 _X86_ X86
            __i386 __i386__
        )
    endif()
endfunction()

# ====================== 工具链初始化 ======================
# 获取指定架构的工具链路径
function(get_toolchain arch output_var)
    # 从统一配置中提取工具链信息
    set(cl "${SDK_ROOT}/${TOOLCHAIN_CONFIG_${arch}_CL_EXE}")
    set(link "${SDK_ROOT}/${TOOLCHAIN_CONFIG_${arch}_LINK_EXE}")
    set(lib "${SDK_ROOT}/${TOOLCHAIN_CONFIG_${arch}_LIB_EXE}")
    set(ml "${SDK_ROOT}/${TOOLCHAIN_CONFIG_${arch}_ML_EXE}")

    # 设置库路径
    set(lib_paths "")
    foreach(path ${TOOLCHAIN_CONFIG_${arch}_LIB_PATHS})
        list(APPEND lib_paths "${SDK_ROOT}/${path}")
    endforeach()

    # 设置返回结构
    set(${output_var}
        CL_EXE ${cl}
        LINK_EXE ${link}
        LIB_EXE ${lib}
        ML_EXE ${ml}
        LIB_PATHS ${lib_paths}
        PARENT_SCOPE
    )
endfunction()

# ====================== 编译选项 ======================
# 获取架构特定的编译标志
function(get_arch_cflags arch output_var)
    if(arch STREQUAL "x64")
        set(flags
            /nologo
            /O2 /GL
            /W4 /WX-
            /D_WIN64 /D_AMD64_ /DAMD64
            /D__x86_64 /D__x86_64__
            /D_UNICODE /DUNICODE
            /DNDEBUG
            /MD
            /EHsc
            /permissive-
            /Zc:wchar_t /Zc:forScope /Zc:inline
            /Gd /TC
        )
    elseif(arch STREQUAL "x86")
        set(flags
            /nologo
            /O1 /GL
            /W3
            /D_WIN32 /D_X86_ /DX86
            /D__i386 /D__i386__
            /D_UNICODE /DUNICODE
            /DNDEBUG
            /MD
            /EHsc
            /permissive-
            /Zc:wchar_t /Zc:forScope /Zc:inline
            /Gd /TC
        )
    endif()

    # 添加包含路径
    foreach(inc ${SDK_INCLUDE_PATHS})
        list(APPEND flags /I"${inc}")
    endforeach()

    set(${output_var} ${flags} PARENT_SCOPE)
endfunction()

# ====================== 链接选项 ======================
# 获取架构特定的链接标志
function(get_arch_ldflags arch type output_var)
    # 基础标志
    set(flags /NOLOGO /LTCG /DYNAMICBASE)

    # EXE特定标志
    if(type STREQUAL "EXE")
        list(APPEND flags /SUBSYSTEM:CONSOLE)
    endif()

    # DLL特定标志
    if(type STREQUAL "DLL")
        list(APPEND flags /DLL)
    endif()

    # 添加库路径
    get_toolchain(${arch} toolchain)
    foreach(path ${toolchain_LIB_PATHS})
        list(APPEND flags /LIBPATH:"${path}")
    endforeach()

    set(${output_var} ${flags} PARENT_SCOPE)
endfunction()

# ====================== 源文件编译 ======================
function(compile_sources target arch sources)
    # 获取工具链和编译标志
    get_toolchain(${arch} toolchain)
    get_arch_cflags(${arch} cflags)

    # 创建对象文件目录
    set(obj_dir "${OBJ_DIR}/${arch}/${target}")
    file(MAKE_DIRECTORY "${obj_dir}")

    # 对象文件列表
    set(obj_files "")

    # 编译每个源文件
    foreach(source_file ${sources})
        # 生成对象文件名
        get_filename_component(name ${source_file} NAME_WE)
        set(obj_file "${obj_dir}/${name}.obj")

        # 确定编译器 (C/C++ 或 MASM)
        if(source_file MATCHES "\\.(asm|s)$")
            set(compiler "${toolchain_ML_EXE}")
            set(compile_flags /c /Cx /Zi)
        else()
            set(compiler "${toolchain_CL_EXE}")
            set(compile_flags ${cflags} /c)
        endif()

        # 添加编译命令
        add_custom_command(
            OUTPUT ${obj_file}
            COMMAND "${compiler}" ${compile_flags}
                    /Fo"${obj_file}"
                    "${source_file}"
            DEPENDS ${source_file}
            COMMENT "编译(${arch}): ${name}"
        )

        list(APPEND obj_files ${obj_file})
    endforeach()

    # 返回对象文件列表
    set(${target}_OBJS ${obj_files} PARENT_SCOPE)
endfunction()

# ====================== 目标构建 ======================
# 构建可执行文件
function(sdk_add_exe target)
    cmake_parse_arguments(ARG "" "ENTRY;ARCH" "SOURCES;LIBS" ${ARGN})
    if(NOT ARG_ARCH)
        set(ARG_ARCH ${SDK_ARCH})
    endif()
    if(NOT ARG_ENTRY)
        set(ARG_ENTRY ${DEFAULT_EXE_ENTRY})
    endif()

    # 添加导航目标
    add_navigation_target(${target}
        SOURCES ${ARG_SOURCES}
        ARCH ${ARG_ARCH}
    )

    # 编译源文件
    compile_sources(${target} ${ARG_ARCH} "${ARG_SOURCES}")

    # 获取工具链和链接标志
    get_toolchain(${ARG_ARCH} toolchain)
    get_arch_ldflags(${ARG_ARCH} "EXE" ldflags)

    # 设置输出路径
    set(output_file "${OUTPUT_DIR}/${ARG_ARCH}/${target}.exe")
    get_filename_component(out_dir "${output_file}" DIRECTORY)
    file(MAKE_DIRECTORY "${out_dir}")

    # 添加构建目标
    add_custom_target(${target} ALL
        COMMAND "${toolchain_LINK_EXE}" ${ldflags}
                /ENTRY:${ARG_ENTRY}
                /OUT:"${output_file}"
                ${${target}_OBJS}
                ${ARG_LIBS}
        DEPENDS ${${target}_OBJS}
        COMMENT "链接EXE(${ARG_ARCH}): ${output_file}"
    )

    # 构建完成消息
    add_custom_command(TARGET ${target} POST_BUILD
        COMMAND ${CMAKE_COMMAND} -E echo "可执行文件构建完成: ${output_file}"
    )

    # 添加清理目标
    add_clean_target(${target} ${output_file} ${${target}_OBJS})
endfunction()

# 构建动态链接库
function(sdk_add_library target)
    cmake_parse_arguments(ARG "" "ENTRY;ARCH;DEF" "SOURCES;LIBS" ${ARGN})
    if(NOT ARG_ARCH)
        set(ARG_ARCH ${SDK_ARCH})
    endif()
    if(NOT ARG_ENTRY)
        set(ARG_ENTRY ${DEFAULT_DLL_ENTRY})
    endif()

    # 添加导航目标
    add_navigation_target(${target}
        SOURCES ${ARG_SOURCES}
        ARCH ${ARG_ARCH}
    )

    # 编译源文件
    compile_sources(${target} ${ARG_ARCH} "${ARG_SOURCES}")

    # 获取工具链和链接标志
    get_toolchain(${ARG_ARCH} toolchain)
    get_arch_ldflags(${ARG_ARCH} "DLL" ldflags)

    # 设置输出路径
    set(output_file "${OUTPUT_DIR}/${ARG_ARCH}/${target}.dll")
    set(implib_file "${OUTPUT_DIR}/${ARG_ARCH}/${target}.lib")
    get_filename_component(out_dir "${output_file}" DIRECTORY)
    file(MAKE_DIRECTORY "${out_dir}")

    # 添加导出库标志
    list(APPEND ldflags /IMPLIB:"${implib_file}")

    # 额外链接选项
    set(extra_opts "")
    if(ARG_DEF)
        list(APPEND extra_opts /DEF:"${ARG_DEF}")
    endif()

    # 添加构建目标
    add_custom_target(${target} ALL
        COMMAND "${toolchain_LINK_EXE}" ${ldflags} ${extra_opts}
                /ENTRY:${ARG_ENTRY}
                /OUT:"${output_file}"
                ${${target}_OBJS}
                ${ARG_LIBS}
        DEPENDS ${${target}_OBJS}
        COMMENT "链接DLL(${ARG_ARCH}): ${output_file}"
    )

    # 构建完成消息
    add_custom_command(TARGET ${target} POST_BUILD
        COMMAND ${CMAKE_COMMAND} -E echo "动态库构建完成: ${output_file}"
    )

    # 添加清理目标
    add_clean_target(${target} ${output_file} ${implib_file} ${${target}_OBJS})
endfunction()

# ====================== 静态库构建 ======================
function(sdk_add_static_library target)
    cmake_parse_arguments(ARG "" "ARCH" "SOURCES" ${ARGN})
    if(NOT ARG_ARCH)
        set(ARG_ARCH ${SDK_ARCH})
    endif()

    # 添加导航目标
    add_navigation_target(${target}
        SOURCES ${ARG_SOURCES}
        ARCH ${ARG_ARCH}
    )

    # 编译源文件
    compile_sources(${target} ${ARG_ARCH} "${ARG_SOURCES}")

    # 获取工具链
    get_toolchain(${ARG_ARCH} toolchain)

    # 设置输出路径
    set(output_file "${OUTPUT_DIR}/${ARG_ARCH}/${target}.lib")
    get_filename_component(out_dir "${output_file}" DIRECTORY)
    file(MAKE_DIRECTORY "${out_dir}")

    # 添加构建目标
    add_custom_target(${target} ALL
        COMMAND "${toolchain_LIB_EXE}" /NOLOGO
                /OUT:"${output_file}"
                ${${target}_OBJS}
        DEPENDS ${${target}_OBJS}
        COMMENT "创建静态库(${ARG_ARCH}): ${output_file}"
    )

    # 构建完成消息
    add_custom_command(TARGET ${target} POST_BUILD
        COMMAND ${CMAKE_COMMAND} -E echo "静态库构建完成: ${output_file}"
    )

    # 添加清理目标
    add_clean_target(${target} ${output_file} ${${target}_OBJS})
endfunction()

# ====================== 清理功能 ======================
# 添加单个目标的清理
function(add_clean_target target)
    set(files_to_clean ${ARGN})
    add_custom_target(clean_${target}
        COMMAND ${CMAKE_COMMAND} -E remove ${files_to_clean}
        COMMENT "清理目标: ${target}"
    )
endfunction()

# 添加全局清理目标
function(add_global_clean_target)
    add_custom_target(clean_all
        COMMAND ${CMAKE_COMMAND} -E remove_directory ${OUTPUT_DIR}
        COMMAND ${CMAKE_COMMAND} -E remove_directory ${OBJ_DIR}
        COMMENT "清理所有构建输出"
    )
endfunction()
