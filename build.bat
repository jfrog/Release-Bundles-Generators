CGO_ENABLED=0 go build -o release-bundle-generator.exe -ldflags '-w -extldflags "-static"' main.go
