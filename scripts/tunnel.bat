@echo off
rem Устойчивый туннель к локальному фронту для Telegram Mini App.
rem Использование:  scripts\tunnel.bat [порт]   (по умолчанию 4174)
rem
rem Что делает:
rem  - поднимает SSH-туннель через localhost.run к 127.0.0.1:<порт>;
rem  - держит соединение живым (ServerAliveInterval);
rem  - при обрыве автоматически переподключается.
rem
rem Важно: если туннель без аккаунта — адрес меняется при каждом переподключении,
rem и его нужно обновлять в Render (WEBAPP_URL). Чтобы адрес был постоянным,
rem добавь свой SSH-ключ на https://admin.localhost.run и используй
rem scripts\tunnel-stable.bat со своим поддоменом.

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
