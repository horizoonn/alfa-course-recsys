@echo off
rem Туннель с ПОСТОЯННЫМ адресом (не меняется при переподключении).
rem
rem Подготовка (один раз):
rem   1) ssh-keygen -t ed25519 -C "alfa-course"
rem      (ключ появится в C:\Users\<ты>\.ssh\id_ed25519.pub)
rem   2) Открой https://admin.localhost.run , войди через GitHub
rem      и добавь содержимое id_ed25519.pub
rem   3) Там же посмотри свой постоянный домен: <username>.lhr.life
rem
rem Запуск: scripts\tunnel-stable.bat <username> [порт]
rem Пример: scripts\tunnel-stable.bat sashauser 4174

setlocal
set SUBDOMAIN=%1
set PORT=%2
if "%SUBDOMAIN%"=="" (
  echo Usage: %~nx0 ^<username^> [port]
  echo Example: %~nx0 sashauser 4174
  exit /b 1
)
if "%PORT%"=="" set PORT=4174

:loop
echo.
echo [tunnel] https://%SUBDOMAIN%.lhr.life -^> 127.0.0.1:%PORT%
ssh -o StrictHostKeyChecking=no ^
    -o ServerAliveInterval=15 ^
    -o ServerAliveCountMax=3 ^
    -o TCPKeepAlive=yes ^
    -o ExitOnForwardFailure=yes ^
    -R %SUBDOMAIN%:80:127.0.0.1:%PORT% nokey@localhost.run

echo [tunnel] connection closed. Reconnecting in 3 seconds...
timeout /t 3 /nobreak >nul
goto loop
