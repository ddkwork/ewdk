# wdk64/wdk.cmake - 完整增强版 (无任何省略)

# 全局配置
set(WDK_ROOT "D:/ewdk/dist/wdk" CACHE PATH "WDK根目录")
set(WDK_VERSION "10.0.26100.0" CACHE STRING "WDK版本")
set(DEFAULT_ENTRY_POINT "DriverEntry" CACHE STRING "默认驱动入口点")

# 主驱动构建函数
function(wdk_add_driver target_name)
    # 解析参数模式
    set(source_patterns "*.c")
    set(include_patterns "*")

    # 如果提供参数，则使用它们
    if(ARGC GREATER 1)
        set(source_patterns ${ARGV1})
        if(ARGC GREATER 2)
            set(include_patterns ${ARGV2})
        endif()
    endif()

    # 收集源文件
    set(driver_sources)
    set(driver_asm_sources)
    set(driver_c_sources)

    # 收集所有文件
    foreach(pattern ${source_patterns})
        if(pattern MATCHES "\\*\\*")
            file(GLOB_RECURSE found_sources
                    LIST_DIRECTORIES false
                    ${pattern}
            )
        else()
            file(GLOB found_sources
                    LIST_DIRECTORIES false
                    ${pattern}
            )
        endif()
        list(APPEND driver_sources ${found_sources})
    endforeach()

    # 分离ASM和C文件
    foreach(file ${driver_sources})
        if(file MATCHES "\\.asm$")
            list(APPEND driver_asm_sources ${file})
        else()
            list(APPEND driver_c_sources ${file})
        endif()
    endforeach()

    # 收集包含目录
    set(driver_includes)
    foreach(pattern ${include_patterns})
        if(pattern MATCHES "\\*\\*")
            file(GLOB_RECURSE found_includes
                    LIST_DIRECTORIES true
                    ${pattern}
            )
        else()
            file(GLOB found_includes
                    LIST_DIRECTORIES true
                    ${pattern}
            )
        endif()

        # 提取目录路径
        foreach(item ${found_includes})
            if(IS_DIRECTORY "${item}")
                list(APPEND driver_includes "${item}")
            else()
                get_filename_component(dir "${item}" DIRECTORY)
                list(APPEND driver_includes "${dir}")
            endif()
        endforeach()
    endforeach()

    # 去重
    list(REMOVE_DUPLICATES driver_sources)
    list(REMOVE_DUPLICATES driver_includes)

    # 回退机制：如果没有源文件，尝试默认模式
    if(NOT driver_sources)
        set(fallback_patterns
                "${target_name}.c"
                "${target_name}.cpp"
                "${target_name}.asm"
                "src/${target_name}.c"
                "src/${target_name}.cpp"
                "src/${target_name}.asm"
        )
        foreach(fallback_pattern ${fallback_patterns})
            file(GLOB fallback_sources ${fallback_pattern})
            if(fallback_sources)
                list(APPEND driver_sources ${fallback_sources})
                break()
            endif()
        endforeach()
    endif()

    # 如果仍然没有源文件，报错
    if(NOT driver_sources)
        message(FATAL_ERROR "无法为驱动 '${target_name}' 找到源文件! "
                "请确认存在匹配的源文件")
    endif()

    # 重新分类源文件
    set(driver_asm_sources)
    set(driver_c_sources)
    foreach(file ${driver_sources})
        if(file MATCHES "\\.asm$")
            list(APPEND driver_asm_sources ${file})
        else()
            list(APPEND driver_c_sources ${file})
        endif()
    endforeach()

    # 如果没有包含目录，添加源文件所在目录
    if(NOT driver_includes)
        foreach(src_file ${driver_sources})
            get_filename_component(src_dir "${src_file}" DIRECTORY)
            list(APPEND driver_includes ${src_dir})
        endforeach()
        list(REMOVE_DUPLICATES driver_includes)
    endif()

    # 设置输出路径
    set(bin_dir "${CMAKE_BINARY_DIR}/bin")
    set(obj_dir "${CMAKE_BINARY_DIR}/obj/${target_name}")
    file(MAKE_DIRECTORY ${bin_dir})
    file(MAKE_DIRECTORY ${obj_dir})

    # 创建禁用运行时检查的文件
    file(WRITE "${CMAKE_BINARY_DIR}/wdkflags.h" "#pragma runtime_checks(\"suc\", off)")

    # 添加WDK核心包含路径
    set(wdk_includes
            ${WDK_ROOT}/Include/${WDK_VERSION}/km
            ${WDK_ROOT}/Include/${WDK_VERSION}/km/crt
            ${WDK_ROOT}/Include/${WDK_VERSION}/shared
            ${WDK_ROOT}/Include/shared
    )

    # 合并所有包含目录
    set(all_includes ${driver_includes} ${wdk_includes})
    list(REMOVE_DUPLICATES all_includes)

    # ==== C/C++ 编译选项 ====
    set(c_compile_flags)
    foreach(dir ${all_includes})
        list(APPEND c_compile_flags /I"${dir}")
    endforeach()

    # 完整编译选项 - 使用 Windows 11 NTDDI 版本
    list(APPEND c_compile_flags
            /nologo /c
            /FI"${CMAKE_BINARY_DIR}/wdkflags.h"
            /W4 /WX   # 警告视为错误
            /GS- /kernel
            /D_AMD64_ /D_WIN64
            /DNTDDI_VERSION=NTDDI_WIN11_ZN  # Windows 11 定义
            /wd4100   # 禁用C4100警告
    )

    # ==== 汇编编译选项 ====
    set(asm_compile_flags)
    foreach(dir ${all_includes})
        list(APPEND asm_compile_flags /I"${dir}")
    endforeach()

    # 汇编器特定选项
    list(APPEND asm_compile_flags
            /nologo
            /c /Cx /Zi
            /D_AMD64_ /D_WIN64
            /DNTDDI_VERSION=NTDDI_WIN11_ZN  # Windows 11 定义
    )

    # 编译所有源文件
    set(obj_files)

    # 编译C/C++源文件
    foreach(source_file ${driver_c_sources})
        # 为每个源文件保持相对路径结构
        file(RELATIVE_PATH rel_path ${CMAKE_CURRENT_SOURCE_DIR} ${source_file})
        set(obj_file "${obj_dir}/${rel_path}.obj")
        get_filename_component(obj_path "${obj_file}" PATH)

        # 确保输出目录存在
        file(MAKE_DIRECTORY "${obj_path}")

        list(APPEND obj_files ${obj_file})

        add_custom_command(
                OUTPUT ${obj_file}
                COMMAND "${WDK_ROOT}/bin/cl.exe"
                ${c_compile_flags}
                /Fo"${obj_file}"
                "${source_file}"
                DEPENDS ${source_file}
                COMMENT "编译: ${rel_path}"
        )
    endforeach()

    # 编译ASM源文件 (使用ml64.exe)
    foreach(asm_file ${driver_asm_sources})
        file(RELATIVE_PATH rel_path ${CMAKE_CURRENT_SOURCE_DIR} ${asm_file})
        set(obj_file "${obj_dir}/${rel_path}.obj")
        get_filename_component(obj_path "${obj_file}" PATH)

        file(MAKE_DIRECTORY "${obj_path}")

        list(APPEND obj_files ${obj_file})

        add_custom_command(
                OUTPUT ${obj_file}
                COMMAND "${WDK_ROOT}/bin/ml64.exe"
                ${asm_compile_flags}
                /Fo"${obj_file}"
                /Wk   # 警告级别设置为最低
                "${asm_file}"
                DEPENDS ${asm_file}
                COMMENT "汇编: ${rel_path}"
        )
    endforeach()

    # 链接驱动
    set(output_file "${bin_dir}/${target_name}.sys")

    add_custom_command(
            OUTPUT ${output_file}
            COMMAND "${WDK_ROOT}/bin/link.exe"
            /nologo /DRIVER
            /SUBSYSTEM:NATIVE /ENTRY:${DEFAULT_ENTRY_POINT}
            /OUT:"${output_file}"
            ${obj_files}
            "${WDK_ROOT}/Lib/${WDK_VERSION}/km/x64/ntoskrnl.lib"
            /ignore:4104  # 忽略ASM文件链接警告
            DEPENDS ${obj_files}
            COMMENT "链接: ${output_file}"
    )

    # 主构建目标
    add_custom_target(${target_name} ALL
            DEPENDS ${output_file}
            COMMENT "构建驱动: ${target_name}.sys"
    )

    # 构建成功提示
    add_custom_command(TARGET ${target_name} POST_BUILD
            COMMAND ${CMAKE_COMMAND} -E echo "驱动构建成功: ${output_file}"
    )

    # 创建专用清理目标
    add_custom_target(clean_${target_name}
            COMMAND ${CMAKE_COMMAND} -E remove ${output_file}
            COMMAND ${CMAKE_COMMAND} -E remove_directory ${obj_dir}
            COMMENT "清理驱动 ${target_name}"
    )

    # 高级导航目标 - 仅包含C/C++文件
    if(driver_c_sources OR driver_asm_sources)
        add_executable(${target_name}_nav ${driver_c_sources})
        set_target_properties(${target_name}_nav PROPERTIES
                EXCLUDE_FROM_ALL TRUE
                EXCLUDE_FROM_DEFAULT_BUILD TRUE
        )
        target_include_directories(${target_name}_nav PRIVATE ${all_includes})
    endif()

    # 打印信息
    message(STATUS "========= 驱动目标: ${target_name} =========")
    message(STATUS "NTDDI_VERSION: NTDDI_WIN11_ZN (Windows 11)")
    message(STATUS "源文件数量: (C/C++:${list_length}, ASM:${driver_asm_sources_count})")
    message(STATUS "包含目录数量: (${all_includes_count})")

    message(STATUS "源文件列表:")
    foreach(src ${driver_sources})
        file(RELATIVE_PATH rel_src ${CMAKE_CURRENT_SOURCE_DIR} ${src})
        message(STATUS "    ${rel_src}")
    endforeach()

    message(STATUS "包含目录列表:")
    foreach(inc ${all_includes})
        file(RELATIVE_PATH rel_inc ${CMAKE_CURRENT_SOURCE_DIR} ${inc})
        message(STATUS "    ${rel_inc}")
    endforeach()

    message(STATUS "输出文件: ${output_file}")
    message(STATUS "=========================================")
endfunction()

# 全局清理目标函数
function(wdk_add_clean_all_target)
    # 收集所有专用清理目标
    set(clean_targets)
    get_property(all_targets DIRECTORY PROPERTY BUILDSYSTEM_TARGETS)
    foreach(target ${all_targets})
        if(target MATCHES "clean_.*")
            list(APPEND clean_targets ${target})
        endif()
    endforeach()

    # 创建全局清理目标
    add_custom_target(driver_clean_all
            COMMAND ${CMAKE_COMMAND} -E rm -rf "${CMAKE_BINARY_DIR}/bin/*.sys"
            COMMAND ${CMAKE_COMMAND} -E rm -rf "${CMAKE_BINARY_DIR}/obj"
            COMMENT "清理所有驱动构建文件"
            DEPENDS ${clean_targets}
    )
endfunction()