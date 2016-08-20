
.PHONY: build
build:
	glide install
	go build -v
