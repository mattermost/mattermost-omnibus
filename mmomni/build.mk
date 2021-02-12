bin:
	@mkdir -p bin


.PHONY: mmomni/vendor
mmomni/vendor:
	cd mmomni; \
	go mod vendor; \
	go mod tidy


bin/mmomni: bin
	@echo "Building mmomni"
	cd mmomni; \
	go build -mod=vendor

	cp mmomni/mmomni bin/mmomni


mmomni-test:
	cd mmomni; \
	go test -v -race ./...


mmomni: bin/mmomni


mmomni-clean:
	@rm -rf bin
