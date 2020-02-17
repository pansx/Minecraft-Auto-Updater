mkdir build
set CGO_ENABLED=0
set GOOS=windows
go build -o build/Updater_windows.exe
set GOOS=linux
go build -o build/Updater_linux
set GOOS=darwin
go build -o build/Updater_darwin