package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strings"
)

var (
	addr             = flag.String("listen-address", ":9031", "The address to listen on for HTTP requests.")
	uwsgi_stats_addr = flag.String("uwsgi-stats-address", "unix://uwsgi.sock", "The address of the uWSGI stats socket.")
)

func parse_uwsgi_stats_addr(address string) (address_type string, addr string, err error) {
	parts := strings.SplitN(address, "://", 2)
	scheme := parts[0]
	path := parts[1]
	switch {
	case scheme == "file":
		address_type = "file"
		addr = path
	case scheme == "fileglob":
		address_type = "fileglob"
		addr = path
	case scheme == "unix":
		address_type = "unix"
		addr = path
	case scheme == "unixglob":
		address_type = "unixglob"
		addr = path
	case scheme == "http":
		address_type = "url"
		addr = address
	default:
		return "", "", fmt.Errorf("Can't parse uwsgi-stats-address: %g", address)
	}

	return address_type, addr, nil
}

func main() {
	flag.Parse()

	address_type, address, err := parse_uwsgi_stats_addr(*uwsgi_stats_addr)
	if err != nil {
		panic(err)
	}

	uwsgi_stats_reader := new_uwsgi_stats_reader(address_type, address)

	collector := NewUwsgiStatsCollector(uwsgi_stats_reader)
	prometheus.MustRegister(collector)

	http.Handle("/metrics", prometheus.Handler())
	fmt.Printf("%s\n", http.ListenAndServe(*addr, nil))
}
