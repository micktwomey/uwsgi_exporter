package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"path"
	"path/filepath"
)

type UwsgiStatsReadResults struct {
	Type       string // e.g. unix, url
	Address    string // e.g. https://.... or /tmp/foo.sock
	Identifier string // host+port or foo.sock
	Error      error
	UwsgiStats UwsgiStats
}

type UwsgiStats struct {
	Version             string
	Listen_Queue        int
	Listen_Queue_Errors int
	Signal_Queue        int
	Load                int
	Pid                 int
	Uid                 int
	Gid                 int
	Cwd                 string
	Sockets             []UwsgiSocket
	Workers             []UwsgiWorker
}

type UwsgiSocket struct {
	Name        string
	Proto       string
	Queue       int
	Max_Queue   int
	Shared      int
	Can_Offload int
}

type UwsgiWorker struct {
	Id             int
	Pid            int
	Accepting      int
	Requests       int
	Delta_Requests int
	Exceptions     int
	Harakiri_Count int
	Signals        int
	Signal_Queue   int
	Status         string
	Rss            int
	Vsz            int
	Running_Time   int
	Last_Spawn     int
	Respawn_Count  int
	Tx             int
	Avg_Rt         int
	Apps           []UwsgiApp
	Cores          []UwsgiCore
}

type UwsgiApp struct {
	Id           int
	Modifier1    int
	Mountpoint   string
	Startup_Time int
	Requests     int
	Exceptions   int
	Chdir        string
}

type UwsgiCore struct {
	Id                int
	Requests          int
	Static_Requests   int
	Routed_Requets    int
	Ofloaded_Requests int
	Write_Errors      int
	Read_Errors       int
	In_Requests       int
	Vars              []string
}

type UwsgiUnixGlobAddr string
type UwsgiUnixAddr string
type UwsgiHttpAddr string
type UwsgiFileAddr string
type UwsgiFileGlobAddr string

type UwsgiStatsReader struct {
	Type    string
	Address string
}

func new_uwsgi_stats_reader(stats_type string, address string) *UwsgiStatsReader {
	return &UwsgiStatsReader{Type: stats_type, Address: address}
}

func read_uwsgi_stats_file(filename UwsgiFileAddr, uwsgi_stats_results *[]UwsgiStatsReadResults) {
	uwsgi_stats := UwsgiStats{}
	results := UwsgiStatsReadResults{
		Type:       "file",
		Address:    string(filename),
		Identifier: path.Base(string(filename)),
		Error:      nil,
		UwsgiStats: uwsgi_stats,
	}
	data, err := ioutil.ReadFile(string(filename))
	if err != nil {
		results.Error = err
	} else {
		results.Error = json.Unmarshal(data, &results.UwsgiStats)
	}
	*uwsgi_stats_results = append(*uwsgi_stats_results, results)
}

func read_uwsgi_stats_fileglob(pattern UwsgiFileGlobAddr, uwsgi_stats_results *[]UwsgiStatsReadResults) {
	matches, err := filepath.Glob(string(pattern))
	if err != nil {
		panic(err)
	}
	for _, filename := range matches {
		read_uwsgi_stats_file(UwsgiFileAddr(filename), uwsgi_stats_results)
	}
}

func read_uwsgi_stats_unix_socket(filename UwsgiUnixAddr, uwsgi_stats_results *[]UwsgiStatsReadResults) {
	uwsgi_stats := UwsgiStats{}
	results := UwsgiStatsReadResults{
		Type:       "unix",
		Address:    string(filename),
		Identifier: path.Base(string(filename)),
		Error:      nil,
		UwsgiStats: uwsgi_stats,
	}
	conn, err := net.Dial("unix", string(filename))
	if err != nil {
		fmt.Errorf("Problem reading %g", filename)
		results.Error = err
	} else {
		defer conn.Close()
		decoder := json.NewDecoder(conn)
		fmt.Errorf("Problem decoding %g from %g", filename, conn)
		results.Error = decoder.Decode(&results.UwsgiStats)
	}
	*uwsgi_stats_results = append(*uwsgi_stats_results, results)
}

func read_uwsgi_stats_unix_glob_socket(glob_pattern UwsgiUnixGlobAddr, uwsgi_stats_results *[]UwsgiStatsReadResults) {
	matches, err := filepath.Glob(string(glob_pattern))
	if err != nil {
		panic(err)
	}
	for _, filename := range matches {
		read_uwsgi_stats_unix_socket(UwsgiUnixAddr(filename), uwsgi_stats_results)
	}

}

func read_uwsgi_stats(reader *UwsgiStatsReader) *[]UwsgiStatsReadResults {
	uwsgi_stats_results := []UwsgiStatsReadResults{}
	switch {
	case reader.Type == "file":
		read_uwsgi_stats_file(UwsgiFileAddr(reader.Address), &uwsgi_stats_results)
	case reader.Type == "fileglob":
		read_uwsgi_stats_fileglob(UwsgiFileGlobAddr(reader.Address), &uwsgi_stats_results)
	case reader.Type == "unix":
		read_uwsgi_stats_unix_socket(UwsgiUnixAddr(reader.Address), &uwsgi_stats_results)
	case reader.Type == "unixglob":
		read_uwsgi_stats_unix_glob_socket(UwsgiUnixGlobAddr(reader.Address), &uwsgi_stats_results)
	default:
		fmt.Errorf("Don't know how to handle %g", reader)
	}
	return &uwsgi_stats_results
}
