@echo off
chcp 65001 > nul
echo 正在设置GOPATH...
set GOPATH=%cd%
echo 完成！
ping -n 3 localhost > nul