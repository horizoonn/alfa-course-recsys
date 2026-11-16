@echo off
rem Устарело: используй node scripts\tunnel.mjs
rem
rem Скрипт tunnel.mjs делает то же самое, но ещё и сам публикует новый адрес
rem в Render (POST /__admin/target), поэтому Mini App не отваливается при
rem переподключении туннеля.
rem
rem Оставлено как резерв: простой SSH-туннель к 127.0.0.1:<порт>.
rem Использование:  scripts\tunnel.bat [порт]   (по умолчанию 4174)
rem
rem Важно: при обрыве адрес меняется, и его нужно вписывать в API_TARGET на Render.

setlocal
set PORT=%1
if "%PORT%"=="" set PORT=4174

:loop
echo.
echo [tunnel] connect 127.0.0.1:%PORT% -^> localhost.run
ssh -o StrictHostKeyChecking=no ^
    -o ServerAliveInterval=15 ^
    -o ServerAliveCountMax=3 ^
    -o TCPKeepAlive=yes ^
    -o ExitOnForwardFailure=yes ^
    -R 80:127.0.0.1:%PORT% nokey@localhost.run

echo [tunnel] connection closed. Reconnecting in 3 seconds...
timeout /t 3 /nobreak >nul
goto loop
