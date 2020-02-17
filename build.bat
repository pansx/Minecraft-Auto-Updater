set CGO_ENABLED=0
set GOOS=windows
go build -o Updater_windows.exe
set GOOS=linux
go build -o Updater_linux
set GOOS=darwin
go build -o Updater_darwin