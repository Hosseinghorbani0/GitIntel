@echo off
setlocal EnableExtensions

set "ROOT=%~dp0"
cd /d "%ROOT%"

if not defined GITINTEL_BACKEND_PORT set "GITINTEL_BACKEND_PORT=8080"
if not defined GITINTEL_FRONTEND_PORT set "GITINTEL_FRONTEND_PORT=5173"

where go >nul 2>nul
if errorlevel 1 (
  echo ERROR: Go was not found in PATH.
  exit /b 1
)

where node >nul 2>nul
if errorlevel 1 (
  echo ERROR: Node.js was not found in PATH.
  exit /b 1
)

where npm >nul 2>nul
if errorlevel 1 (
  echo ERROR: npm was not found in PATH.
  exit /b 1
)

if "%GITINTEL_BACKEND_PORT%"=="%GITINTEL_FRONTEND_PORT%" (
  echo ERROR: Backend and frontend ports must be different.
  exit /b 1
)

call :check_port "%GITINTEL_BACKEND_PORT%" "Backend"
if errorlevel 1 exit /b 1
call :check_port "%GITINTEL_FRONTEND_PORT%" "Frontend"
if errorlevel 1 exit /b 1

if not exist "frontend\node_modules" (
  echo Installing frontend dependencies from package-lock.json...
  pushd "frontend"
  call npm ci
  if errorlevel 1 (
    popd
    echo ERROR: Frontend dependency installation failed.
    exit /b 1
  )
  popd
)

echo Starting GitIntel backend on http://127.0.0.1:%GITINTEL_BACKEND_PORT% ...
start "GitIntel Backend" cmd.exe /k "cd /d ""%ROOT%backend"" && set PORT=%GITINTEL_BACKEND_PORT% && go run ./cmd/server"
call :wait_backend "%GITINTEL_BACKEND_PORT%"
if errorlevel 1 (
  echo ERROR: Backend did not become healthy within 30 seconds. Check the GitIntel Backend console.
  exit /b 1
)
echo Backend ready: http://127.0.0.1:%GITINTEL_BACKEND_PORT%/api/health

echo Starting GitIntel frontend on http://127.0.0.1:%GITINTEL_FRONTEND_PORT% ...
start "GitIntel Frontend" cmd.exe /k "cd /d ""%ROOT%frontend"" && set GITINTEL_API_URL=http://127.0.0.1:%GITINTEL_BACKEND_PORT% && set GITINTEL_FRONTEND_PORT=%GITINTEL_FRONTEND_PORT% && npm run dev -- --host 127.0.0.1 --strictPort"
call :wait_frontend "%GITINTEL_FRONTEND_PORT%"
if errorlevel 1 (
  echo ERROR: Frontend did not become ready within 30 seconds. Check the GitIntel Frontend console.
  exit /b 1
)

echo Frontend ready: http://127.0.0.1:%GITINTEL_FRONTEND_PORT%/
echo Backend API:    http://127.0.0.1:%GITINTEL_BACKEND_PORT%/api
start "" "http://127.0.0.1:%GITINTEL_FRONTEND_PORT%/"
exit /b 0

:check_port
powershell.exe -NoProfile -Command "$p=%~1; if ($p -notmatch '^\d+$' -or [int]$p -lt 1 -or [int]$p -gt 65535) { exit 2 }; if (Get-NetTCPConnection -State Listen -LocalPort ([int]$p) -ErrorAction SilentlyContinue) { exit 1 }" >nul 2>nul
if errorlevel 2 (
  echo ERROR: %~2 port %~1 is invalid.
  exit /b 1
)
if errorlevel 1 (
  echo ERROR: %~2 port %~1 is already in use. No process was stopped.
  echo Close the owning application or choose another port, for example:
  if /i "%~2"=="Backend" echo   set GITINTEL_BACKEND_PORT=8081
  if /i "%~2"=="Frontend" echo   set GITINTEL_FRONTEND_PORT=5174
  exit /b 1
)
exit /b 0

:wait_backend
powershell.exe -NoProfile -Command "$url='http://127.0.0.1:%~1/api/health'; $deadline=(Get-Date).AddSeconds(30); while ((Get-Date) -lt $deadline) { try { $r=Invoke-RestMethod -Uri $url -TimeoutSec 1; if ($r.success -and $r.data.status -eq 'ok') { exit 0 } } catch {}; [System.Threading.Thread]::Sleep(500) }; exit 1" >nul 2>nul
exit /b %errorlevel%

:wait_frontend
powershell.exe -NoProfile -Command "$url='http://127.0.0.1:%~1/'; $deadline=(Get-Date).AddSeconds(30); while ((Get-Date) -lt $deadline) { try { $r=Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 1; if ($r.StatusCode -eq 200) { exit 0 } } catch {}; [System.Threading.Thread]::Sleep(500) }; exit 1" >nul 2>nul
exit /b %errorlevel%