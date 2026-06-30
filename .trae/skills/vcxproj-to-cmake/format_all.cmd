@echo off
for /r %%i in (*.c *.cpp *.h *.hpp) do clang-format -i -style=file "%%i"
echo Done.