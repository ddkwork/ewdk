cmake -B Release -G "Ninja" -DCMAKE_BUILD_TYPE=Release . && cmake --build Release --config Release
cmake -B Debug -G "Ninja" -DCMAKE_BUILD_TYPE=Debug . && cmake --build Debug --config Debug
