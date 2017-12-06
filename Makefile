
.PHONY: build
build:
	go build -v

.PHONY: build-deb
build-deb:
	go build -v -o uwsgi-exporter
	docker build -t amitsaha/fpm -f Dockerfile.fpm .
	docker run -v ${CURDIR}:/workspace -ti amitsaha/fpm fpm -s dir -t deb -n uwsgi-exporter -v 0.1 uwsgi-exporter=/usr/bin/

.PHONY: glide-install
glide-install:
	glide install --strip-vcs

.PHONY: glide-update
glide-update:
	glide update --strip-vcs
