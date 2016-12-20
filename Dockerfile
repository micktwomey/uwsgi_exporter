FROM golang:1.7.4
COPY . /go/src/uwsgi_exporter
WORKDIR /go/src/uwsgi_exporter
RUN go get -d -v && go build -o uwsgi_exporter && go install && cp /go/bin/uwsgi_exporter /usr/local/bin/uwsgi_exporter
VOLUME /tmp/uwsgi_stats.sock
EXPOSE 9131
CMD ["uwsgi_exporter", "-listen-address", ":9131", "-uwsgi-stats-address", "unix://tmp/uwsgi_stats.sock"]
