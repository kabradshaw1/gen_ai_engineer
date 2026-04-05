# Run this script as Administrator
# Creates scheduled tasks for Minikube Tunnel and Ollama Serve

$ErrorActionPreference = "Stop"

# 1. Minikube Tunnel — runs as SYSTEM with highest privileges
Write-Host "Creating 'Minikube Tunnel' scheduled task..." -ForegroundColor Cyan
$action = New-ScheduledTaskAction -Execute "C:\Program Files\Kubernetes\Minikube\minikube.exe" -Argument "tunnel"
$trigger = New-ScheduledTaskTrigger -AtStartup
$trigger.Delay = "PT30S"
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)
$principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -RunLevel Highest
Register-ScheduledTask -TaskName "Minikube Tunnel" -Action $action -Trigger $trigger -Settings $settings -Principal $principal -Description "Keeps minikube tunnel running for Ingress on localhost:80" -Force
Write-Host "  Done." -ForegroundColor Green

# 2. Ollama Serve — runs as current user with OLLAMA_HOST=0.0.0.0
Write-Host "Creating 'Ollama Serve' scheduled task..." -ForegroundColor Cyan
$action = New-ScheduledTaskAction -Execute "cmd.exe" -Argument '/c set OLLAMA_HOST=0.0.0.0 && set OLLAMA_ORIGINS=* && "C:\Users\PC\AppData\Local\Programs\Ollama\ollama.exe" serve'
$trigger = New-ScheduledTaskTrigger -AtStartup
$trigger.Delay = "PT10S"
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)
$principal = New-ScheduledTaskPrincipal -UserId "PC" -RunLevel Highest -LogonType S4U
Register-ScheduledTask -TaskName "Ollama Serve" -Action $action -Trigger $trigger -Settings $settings -Principal $principal -Description "Runs Ollama with OLLAMA_HOST=0.0.0.0 for Minikube access" -Force
Write-Host "  Done." -ForegroundColor Green

Write-Host ""
Write-Host "Both tasks created. They will run automatically on startup." -ForegroundColor Yellow
Write-Host "To test now: schtasks /run /tn 'Minikube Tunnel'" -ForegroundColor Yellow
