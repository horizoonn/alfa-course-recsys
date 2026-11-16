$ErrorActionPreference = 'Stop'
$tunnelScript = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot 'tunnel.mjs'))
$tunnelPattern = [regex]::Escape($tunnelScript)
$processes = Get-CimInstance Win32_Process
$tunnels = $processes | Where-Object {
    $_.Name -eq 'node.exe' -and $_.CommandLine -match $tunnelPattern
}
foreach ($tunnel in $tunnels) {
    $processes | Where-Object {
        $_.Name -eq 'ssh.exe' -and $_.ParentProcessId -eq $tunnel.ProcessId
    } | ForEach-Object { Stop-Process -Id $_.ProcessId -ErrorAction SilentlyContinue }
    Stop-Process -Id $tunnel.ProcessId -ErrorAction SilentlyContinue
    Write-Host "Stopped project tunnel $($tunnel.ProcessId)"
}
if (-not $tunnels) { Write-Host 'Project tunnel is not running.' }
