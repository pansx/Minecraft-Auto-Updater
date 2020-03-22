mkdir build
set CGO_ENABLED=0
set GOOS=windows
go build -o build/MAU_windows.exe
set GOOS=linux
go build -o build/MAU_linux
set GOOS=darwin
go build -o build/MAU_darwin