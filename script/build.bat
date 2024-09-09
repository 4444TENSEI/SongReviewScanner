@echo off
set GOOS=windows
set GOARCH=amd64

cd .\asset\icon
windres -o Favicon.syso -i Favicon.rc
move /Y Favicon.syso ..\..

cd ..\..
go build -o SongReviewScanner.exe
move /Y SongReviewScanner.exe .\build
