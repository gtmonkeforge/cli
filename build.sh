mkdir -p bin/
GOOS=linux GOARCH=amd64 go build -o bin/mforge-linux-x64 src/*
GOOS=windows GOARCH=amd64 go build -o bin/mforge-win-x64.exe src/*