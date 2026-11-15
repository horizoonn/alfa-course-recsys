' Starts the tunnel in the background without a console window.
' Closing the terminal no longer kills the tunnel.
'
' Usage: double-click this file, or run:  wscript scripts\tunnel-daemon.vbs
' Log:   tmp\tunnel.log
' Stop:  scripts\tunnel-stop.bat

Set fso = CreateObject("Scripting.FileSystemObject")
scriptDir = fso.GetParentFolderName(WScript.ScriptFullName)
root = fso.GetParentFolderName(scriptDir)

If Not fso.FolderExists(root & "\tmp") Then
  fso.CreateFolder root & "\tmp"
End If

Set shell = CreateObject("WScript.Shell")
shell.CurrentDirectory = root
Set service = GetObject("winmgmts:\\.\root\cimv2")
For Each process In service.ExecQuery("SELECT CommandLine FROM Win32_Process WHERE Name='node.exe'")
  If Not IsNull(process.CommandLine) Then
    If InStr(1, process.CommandLine, root & "\scripts\tunnel.mjs", 1) > 0 Then
      WScript.Quit 0
    End If
  End If
Next
shell.Run "cmd /c node """ & root & "\scripts\tunnel.mjs"" >> """ & root & "\tmp\tunnel.log"" 2>&1", 0, False
