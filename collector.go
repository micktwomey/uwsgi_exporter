package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/oleiade/reflections.v1"
	"strconv"
	"strings"
)

type UwsgiStatsCollector struct {
	Reader           *UwsgiStatsReader
	ReadsDesc        *prometheus.Desc
	ErrorDesc        *prometheus.Desc
	UwsgiStats       []Stat
	UwsgiSocketStats []Stat
	UwsgiWorkerStats []Stat
	UwsgiAppStats    []Stat
	UwsgiCoreStats   []Stat
}

type Stat struct {
	Name           string
	PrometheusType prometheus.ValueType
	Desc           *prometheus.Desc
}

func (u *UwsgiStatsCollector) GatherUwsgiStats() *[]UwsgiStatsReadResults {
	uwsgi_stats_results := read_uwsgi_stats(u.Reader)
	return uwsgi_stats_results
}

func (u *UwsgiStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- u.ReadsDesc
	ch <- u.ErrorDesc
	for _, stat := range u.UwsgiStats {
		ch <- stat.Desc
	}
	for _, stat := range u.UwsgiSocketStats {
		ch <- stat.Desc
	}
	for _, stat := range u.UwsgiWorkerStats {
		ch <- stat.Desc
	}
	for _, stat := range u.UwsgiAppStats {
		ch <- stat.Desc
	}
	for _, stat := range u.UwsgiCoreStats {
		ch <- stat.Desc
	}
}

func NewStatMetric(stat *Stat, value int, results *UwsgiStatsReadResults) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		stat.Desc,
		stat.PrometheusType,
		float64(value),
		results.Type,
		results.Address,
		results.Identifier,
	)
}

func NewSocketStatMetric(stat *Stat, value int, socket *UwsgiSocket, results *UwsgiStatsReadResults) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		stat.Desc,
		stat.PrometheusType,
		float64(value),
		results.Type,
		results.Address,
		results.Identifier,
		socket.Name,
		socket.Proto,
	)
}

func NewWorkerStatMetric(stat *Stat, value int, worker *UwsgiWorker, results *UwsgiStatsReadResults) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		stat.Desc,
		stat.PrometheusType,
		float64(value),
		results.Type,
		results.Address,
		results.Identifier,
		strconv.Itoa(worker.Id),
		worker.Status,
	)
}

func NewAppStatMetric(stat *Stat, value int, app *UwsgiApp, worker *UwsgiWorker, results *UwsgiStatsReadResults) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		stat.Desc,
		stat.PrometheusType,
		float64(value),
		results.Type,
		results.Address,
		results.Identifier,
		strconv.Itoa(worker.Id),
		worker.Status,
		strconv.Itoa(app.Id),
		app.Mountpoint,
		app.Chdir,
	)
}

func NewCoreStatMetric(stat *Stat, value int, core *UwsgiCore, worker *UwsgiWorker, results *UwsgiStatsReadResults) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		stat.Desc,
		stat.PrometheusType,
		float64(value),
		results.Type,
		results.Address,
		results.Identifier,
		strconv.Itoa(worker.Id),
		worker.Status,
		strconv.Itoa(core.Id),
	)
}

func (u *UwsgiStatsCollector) Collect(ch chan<- prometheus.Metric) {
	uwsgi_stats_results := u.GatherUwsgiStats()
	for _, results := range *uwsgi_stats_results {
		if results.Error != nil {
			ch <- prometheus.MustNewConstMetric(
				u.ErrorDesc,
				prometheus.CounterValue,
				float64(1),
				results.Type,
				results.Address,
				results.Identifier,
			)
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			u.ReadsDesc,
			prometheus.CounterValue,
			float64(1),
			results.Type,
			results.Address,
			results.Identifier,
			results.UwsgiStats.Version,
		)
		for _, stat := range u.UwsgiStats {
			value, err := reflections.GetField(results.UwsgiStats, stat.Name)
			if err != nil {
				panic(err)
			}
			ch <- NewStatMetric(&stat, value.(int), &results)
		}
		for _, stat := range u.UwsgiSocketStats {
			for _, uwsgi_socket := range results.UwsgiStats.Sockets {
				value, err := reflections.GetField(uwsgi_socket, stat.Name)
				if err != nil {
					panic(err)
				}
				ch <- NewSocketStatMetric(&stat, value.(int), &uwsgi_socket, &results)
			}
		}
		for _, stat := range u.UwsgiWorkerStats {
			for _, uwsgi_worker := range results.UwsgiStats.Workers {
				value, err := reflections.GetField(uwsgi_worker, stat.Name)
				if err != nil {
					panic(err)
				}
				ch <- NewWorkerStatMetric(&stat, value.(int), &uwsgi_worker, &results)
			}
		}
		for _, stat := range u.UwsgiAppStats {
			for _, worker := range results.UwsgiStats.Workers {
				for _, app := range worker.Apps {
					value, err := reflections.GetField(app, stat.Name)
					if err != nil {
						panic(err)
					}
					ch <- NewAppStatMetric(&stat, value.(int), &app, &worker, &results)
				}
			}
		}
		for _, stat := range u.UwsgiCoreStats {
			for _, worker := range results.UwsgiStats.Workers {
				for _, core := range worker.Cores {
					value, err := reflections.GetField(core, stat.Name)
					if err != nil {
						panic(err)
					}
					ch <- NewCoreStatMetric(&stat, value.(int), &core, &worker, &results)
				}
			}
		}
	}
}

func NewUwsgiGaugeStat(name string, description string, prefix string, label_names *[]string) Stat {
	prometheus_name := prefix + strings.ToLower(name)
	return Stat{
		name,
		prometheus.GaugeValue,
		prometheus.NewDesc(
			prometheus_name,
			description,
			*label_names,
			prometheus.Labels{},
		),
	}
}

func NewUwsgiCounterStat(name string, description string, prefix string, suffix string, label_names *[]string) Stat {
	prometheus_name := prefix + strings.ToLower(name) + suffix
	return Stat{
		name,
		prometheus.CounterValue,
		prometheus.NewDesc(
			prometheus_name,
			description,
			*label_names,
			prometheus.Labels{},
		),
	}
}

func NewUwsgiStats() []Stat {
	prefix := "uwsgi_stats_"
	label_names := []string{"type", "uwsgi_stats_address", "identifier"}
	return []Stat{
		NewUwsgiGaugeStat("Listen_Queue", "Length of listen queue.", prefix, &label_names),
		NewUwsgiGaugeStat("Listen_Queue_Errors", "Number of listen queue errors.", prefix, &label_names),
		NewUwsgiGaugeStat("Signal_Queue", "Length of signal queue.", prefix, &label_names),
		NewUwsgiGaugeStat("Load", "Load.", prefix, &label_names),
	}
}

func NewUwsgiSocketStats() []Stat {
	prefix := "uwsgi_stats_socket_"
	label_names := []string{"type", "uwsgi_stats_address", "identifier", "name", "proto"}
	return []Stat{
		NewUwsgiGaugeStat("Queue", "Length of socket queue.", prefix, &label_names),
		NewUwsgiGaugeStat("Max_Queue", "Make length of socket queue.", prefix, &label_names),
		NewUwsgiGaugeStat("Shared", "Shared.", prefix, &label_names),
		NewUwsgiGaugeStat("Can_Offload", "Can socket offload?", prefix, &label_names),
	}
}

func NewUwsgiWorkerStats() []Stat {
	prefix := "uwsgi_stats_worker_"
	suffix := "_total"
	label_names := []string{"type", "uwsgi_stats_address", "identifier", "worker_id", "status"}
	return []Stat{
		NewUwsgiGaugeStat("Accepting", "Is this worker accepting requests?.", prefix, &label_names),
		NewUwsgiCounterStat("Requests", "Number of requests.", prefix, suffix, &label_names),
		NewUwsgiCounterStat("Delta_Requests", "Number of delta requests.", prefix, suffix, &label_names),
		NewUwsgiCounterStat("Exceptions", "Number of exceptions.", prefix, suffix, &label_names),
		NewUwsgiCounterStat("Harakiri_Count", "Number of harakiri attempts.", prefix, suffix, &label_names),
		NewUwsgiCounterStat("Signals", "Number of signals.", prefix, suffix, &label_names),
		NewUwsgiGaugeStat("Signal_Queue", "Length of signal queue.", prefix, &label_names),
		NewUwsgiGaugeStat("Rss", "Worker RSS bytes.", prefix, &label_names),
		NewUwsgiGaugeStat("Vsz", "Worker VSZ bytes.", prefix, &label_names),
		NewUwsgiGaugeStat("Running_Time", "Worker running time.", prefix, &label_names),
		NewUwsgiGaugeStat("Last_Spawn", "Last worker respawn time.", prefix, &label_names),
		NewUwsgiCounterStat("Respawn_Count", "Worker respawn count.", prefix, suffix, &label_names),
		NewUwsgiGaugeStat("Tx", "Worker transmitted bytes.", prefix, &label_names),
		NewUwsgiGaugeStat("Avg_Rt", "Worker average RT.", prefix, &label_names),
	}
}

func NewUwsgiAppStats() []Stat {
	prefix := "uwsgi_stats_worker_app_"
	suffix := "_total"
	label_names := []string{"type", "uwsgi_stats_address", "identifier", "worker_id", "status", "app_id", "mountpoint", "chdir"}
	return []Stat{
		NewUwsgiGaugeStat("Startup_Time", "How long this app took to start.", prefix, &label_names),
		NewUwsgiCounterStat("Requests", "Number of requests.", prefix, suffix, &label_names),
		NewUwsgiCounterStat("Exceptions", "Number of exceptions.", prefix, suffix, &label_names),
	}
}

func NewUwsgiCoreStats() []Stat {
	prefix := "uwsgi_stats_worker_core_"
	suffix := "_total"
	label_names := []string{"type", "uwsgi_stats_address", "identifier", "worker_id", "status", "core_id"}
	return []Stat{
		NewUwsgiCounterStat("Requests", "Number of requests.", prefix, suffix, &label_names),
		NewUwsgiCounterStat("Static_Requests", "Number of static requests.", prefix, suffix, &label_names),
		NewUwsgiCounterStat("Routed_Requets", "Number of routed requests.", prefix, suffix, &label_names),
		NewUwsgiCounterStat("Ofloaded_Requests", "Number of requests offloaded to threads.", prefix, suffix, &label_names),
		NewUwsgiCounterStat("Write_Errors", "Number of write errors.", prefix, suffix, &label_names),
		NewUwsgiCounterStat("Read_Errors", "Number of read errors.", prefix, suffix, &label_names),
		NewUwsgiCounterStat("In_Requests", "Number of requests in.", prefix, suffix, &label_names),
	}
}

func NewUwsgiStatsCollector(reader *UwsgiStatsReader) *UwsgiStatsCollector {

	return &UwsgiStatsCollector{
		Reader: reader,
		ReadsDesc: prometheus.NewDesc(
			"uwsgi_stats_scrapes_total",
			"Number of times stats are scraped",
			[]string{"type", "uwsgi_stats_address", "identifier", "uwsgi_version"},
			prometheus.Labels{},
		),
		ErrorDesc: prometheus.NewDesc(
			"uwsgi_stats_read_error_total",
			"Problems reading stats",
			[]string{"type", "uwsgi_stats_address", "identifier"},
			prometheus.Labels{},
		),
		UwsgiStats:       NewUwsgiStats(),
		UwsgiSocketStats: NewUwsgiSocketStats(),
		UwsgiWorkerStats: NewUwsgiWorkerStats(),
		UwsgiAppStats:    NewUwsgiAppStats(),
		UwsgiCoreStats:   NewUwsgiCoreStats(),
	}
}
