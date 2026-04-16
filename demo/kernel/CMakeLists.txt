cmake_minimum_required(VERSION 3.16)
project(KernelDriverDemo)

wdk_add_driver(KernelDriverDemo main.c)
