cmake_minimum_required(VERSION 3.16)
if(NOT COMMAND um_dll)
    include("C:/Program Files/CMake/bin/ewdk.cmake")
endif()

project(qt6_bindings C CXX)

um_dll(qt6_bindings
    ${CMAKE_CURRENT_SOURCE_DIR}/cflags_windowsqtstatic.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/sync_stubs.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/qt_version_tag_stub.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/callback_stubs.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/gen_qcoreapplication.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/gen_qapplication.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/gen_qobject.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/gen_qpushbutton.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/gen_qtreewidget.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/gen_qtimer.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/gen_qwidget.cpp
)
