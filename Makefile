BIN_M2I=mpo2img
BIN_I2M=img2mpo

HEAD=$(shell git describe --tags 2> /dev/null  || git rev-parse --short HEAD)

all: build

build: darwin_amd64 darwin_arm64 linux_amd64 windows_amd64

clean:
	-rm -f $(BIN)
	-rm -rf release

darwin_amd64:
	env GOOS=darwin GOARCH=amd64 go clean -i ./cmd/$(BIN_M2I)
	env GOOS=darwin GOARCH=amd64 go build -o release/darwin_amd64/$(BIN_M2I) ./cmd/$(BIN_M2I)
	env GOOS=darwin GOARCH=amd64 go clean -i ./cmd/$(BIN_I2M)
	env GOOS=darwin GOARCH=amd64 go build -o release/darwin_amd64/$(BIN_I2M) ./cmd/$(BIN_I2M)

darwin_arm64:
	env GOOS=darwin GOARCH=arm64 go clean -i ./cmd/$(BIN_M2I)
	env GOOS=darwin GOARCH=arm64 go build -o release/darwin_arm64/$(BIN_M2I) ./cmd/$(BIN_M2I)
	env GOOS=darwin GOARCH=arm64 go clean -i ./cmd/$(BIN_I2M)
	env GOOS=darwin GOARCH=arm64 go build -o release/darwin_arm64/$(BIN_I2M) ./cmd/$(BIN_I2M)

linux_amd64:
	env GOOS=linux GOARCH=amd64 go clean -i ./cmd/$(BIN_M2I)
	env GOOS=linux GOARCH=amd64 go build -o release/linux_amd64/$(BIN_M2I) ./cmd/$(BIN_M2I)
	env GOOS=linux GOARCH=amd64 go clean -i ./cmd/$(BIN_I2M)
	env GOOS=linux GOARCH=amd64 go build -o release/linux_amd64/$(BIN_I2M) ./cmd/$(BIN_I2M)

windows_amd64:
	env GOOS=windows GOARCH=amd64 go clean -i ./cmd/$(BIN_M2I)
	env GOOS=windows GOARCH=amd64 go build -o release/windows_amd64/$(BIN_M2I).exe ./cmd/$(BIN_M2I)
	env GOOS=windows GOARCH=amd64 go clean -i ./cmd/$(BIN_I2M)
	env GOOS=windows GOARCH=amd64 go build -o release/windows_amd64/$(BIN_I2M).exe ./cmd/$(BIN_I2M)

.PHONY: release
release: clean build
	zip -9 release/$(BIN_M2I).darwin_amd64.$(HEAD).zip release/darwin_amd64/$(BIN_M2I)
	zip -9 release/$(BIN_M2I).darwin_arm64.$(HEAD).zip release/darwin_arm64/$(BIN_M2I)
	zip -9 release/$(BIN_M2I).linux_amd64.$(HEAD).zip release/linux_amd64/$(BIN_M2I)
	zip -9 release/$(BIN_M2I).windows_amd64.$(HEAD).zip release/windows_amd64/$(BIN_M2I).exe
	zip -9 release/$(BIN_I2M).darwin_amd64.$(HEAD).zip release/darwin_amd64/$(BIN_I2M)
	zip -9 release/$(BIN_I2M).darwin_arm64.$(HEAD).zip release/darwin_arm64/$(BIN_I2M)
	zip -9 release/$(BIN_I2M).linux_amd64.$(HEAD).zip release/linux_amd64/$(BIN_I2M)
	zip -9 release/$(BIN_I2M).windows_amd64.$(HEAD).zip release/windows_amd64/$(BIN_I2M).exe
