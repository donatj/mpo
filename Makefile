BIN_M2I=mpo2img
BIN_I2M=img2mpo
CODESIGN_IDENTITY=Developer ID Application: JESSE GORDON DONAT (NBWN497MH2)
NOTARY_PROFILE=notarytool-profile

HEAD=$(shell git describe --tags 2> /dev/null  || git rev-parse --short HEAD)

all: clean test build

.PHONY: test
test:
	go test ./...

clean:
	-rm -f $(BIN)
	-rm -rf release dist

release/darwin_amd64/$(BIN_M2I) release/darwin_amd64/$(BIN_I2M):
	env GOOS=darwin GOARCH=amd64 go clean -i ./cmd/$(BIN_M2I)
	env GOOS=darwin GOARCH=amd64 go build -o release/darwin_amd64/$(BIN_M2I) ./cmd/$(BIN_M2I)
	env GOOS=darwin GOARCH=amd64 go clean -i ./cmd/$(BIN_I2M)
	env GOOS=darwin GOARCH=amd64 go build -o release/darwin_amd64/$(BIN_I2M) ./cmd/$(BIN_I2M)

release/darwin_arm64/$(BIN_M2I) release/darwin_arm64/$(BIN_I2M):
	env GOOS=darwin GOARCH=arm64 go clean -i ./cmd/$(BIN_M2I)
	env GOOS=darwin GOARCH=arm64 go build -o release/darwin_arm64/$(BIN_M2I) ./cmd/$(BIN_M2I)
	env GOOS=darwin GOARCH=arm64 go clean -i ./cmd/$(BIN_I2M)
	env GOOS=darwin GOARCH=arm64 go build -o release/darwin_arm64/$(BIN_I2M) ./cmd/$(BIN_I2M)

release/darwin_universal/$(BIN_M2I) release/darwin_universal/$(BIN_I2M): release/darwin_amd64/$(BIN_M2I) release/darwin_arm64/$(BIN_M2I) release/darwin_amd64/$(BIN_I2M) release/darwin_arm64/$(BIN_I2M)
	mkdir release/darwin_universal
	lipo -create -output release/darwin_universal/$(BIN_M2I) release/darwin_amd64/$(BIN_M2I) release/darwin_arm64/$(BIN_M2I)
	lipo -create -output release/darwin_universal/$(BIN_I2M) release/darwin_amd64/$(BIN_I2M) release/darwin_arm64/$(BIN_I2M)

release/linux_amd64/$(BIN_M2I) release/linux_amd64/$(BIN_I2M):
	env GOOS=linux GOARCH=amd64 go clean -i ./cmd/$(BIN_M2I)
	env GOOS=linux GOARCH=amd64 go build -o release/linux_amd64/$(BIN_M2I) ./cmd/$(BIN_M2I)
	env GOOS=linux GOARCH=amd64 go clean -i ./cmd/$(BIN_I2M)
	env GOOS=linux GOARCH=amd64 go build -o release/linux_amd64/$(BIN_I2M) ./cmd/$(BIN_I2M)

release/windows_amd64/$(BIN_M2I).exe release/windows_amd64/$(BIN_I2M).exe:
	env GOOS=windows GOARCH=amd64 go clean -i ./cmd/$(BIN_M2I)
	env GOOS=windows GOARCH=amd64 go build -o release/windows_amd64/$(BIN_M2I).exe ./cmd/$(BIN_M2I)
	env GOOS=windows GOARCH=amd64 go clean -i ./cmd/$(BIN_I2M)
	env GOOS=windows GOARCH=amd64 go build -o release/windows_amd64/$(BIN_I2M).exe ./cmd/$(BIN_I2M)

.PHONY: sign
sign: build
	codesign \
		--force \
		--timestamp \
		--options runtime \
		--sign "$(CODESIGN_IDENTITY)" \
		release/darwin_universal/$(BIN_M2I)

	codesign --verify --strict --verbose=4 release/darwin_universal/$(BIN_M2I)

	codesign \
		--force \
		--timestamp \
		--options runtime \
		--sign "$(CODESIGN_IDENTITY)" \
		release/darwin_universal/$(BIN_I2M)

	codesign --verify --strict --verbose=4 release/darwin_universal/$(BIN_I2M)

.PHONY: package
package: sign
	mkdir -p dist
	ditto -c -k --keepParent release/darwin_universal/$(BIN_M2I) dist/$(BIN_M2I).darwin_universal.zip
	ditto -c -k --keepParent release/darwin_universal/$(BIN_I2M) dist/$(BIN_I2M).darwin_universal.zip

.PHONY: notarize
notarize: package
	xcrun notarytool submit dist/$(BIN_M2I).darwin_universal.zip \
		--keychain-profile "$(NOTARY_PROFILE)" \
		--wait

	xcrun notarytool submit dist/$(BIN_I2M).darwin_universal.zip \
		--keychain-profile "$(NOTARY_PROFILE)" \
		--wait

.PHONY: build
build: release/darwin_universal/$(BIN_M2I) release/darwin_universal/$(BIN_I2M) release/linux_amd64/$(BIN_M2I) release/linux_amd64/$(BIN_I2M) release/windows_amd64/$(BIN_M2I).exe release/windows_amd64/$(BIN_I2M).exe

.PHONY: release
release: clean build
	$(MAKE) notarize
	zip -9 dist/$(BIN_M2I).linux_amd64.$(HEAD).zip release/linux_amd64/$(BIN_M2I)
	zip -9 dist/$(BIN_M2I).windows_amd64.$(HEAD).zip release/windows_amd64/$(BIN_M2I).exe
	zip -9 dist/$(BIN_I2M).linux_amd64.$(HEAD).zip release/linux_amd64/$(BIN_I2M)
	zip -9 dist/$(BIN_I2M).windows_amd64.$(HEAD).zip release/windows_amd64/$(BIN_I2M).exe
