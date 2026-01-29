@echo off
echo Starting Go-Gateway...

REM 检查是否已构建可执行文件
if exist gateway.exe (
    echo Using existing gateway.exe
    gateway.exe
) else (
    echo Building and running gateway...
    go run .
)

pause