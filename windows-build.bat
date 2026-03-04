@echo off
echo [*] Building React frontend (Single File)...
cd ui
call npm install
call npm run build
cd ..

echo [*] Copying template...
if not exist "pkg\generator\template" mkdir "pkg\generator\template"
copy /Y "ui\dist\index.html" "pkg\generator\template\index.html"

echo [*] Compiling Go binary for Windows...
if not exist "bin" mkdir "bin"
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-w -s" -o bin\git-recon.exe .\cmd\git-recon

echo [+] Build complete. Executable is in the bin\ folder.