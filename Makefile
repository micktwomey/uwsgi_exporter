
.PHONY: build
build:
	go build -v

.PHONY: glide-install
glide-install:
	glide install --strip-vcs

.PHONY: glide-update
glide-update:
	glide update --strip-vcs
