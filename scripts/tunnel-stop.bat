@echo off
rem Stops only this project's tunnel and its own ssh child.
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0tunnel-stop.ps1"
