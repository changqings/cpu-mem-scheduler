package main

import (
	"cpu-mem-scheduler/handler"
	"log"
	"log/slog"
)

var (
	// change it for your own prometheus url
	// and node-exporter and kube-state-metrics have installed
	prometheusUrl = "http://prometheus.xx"
)

func main() {
	slog.Info("start run k8s scheduler", "name", handler.SchedulerName)
	s := handler.NewScheduler(prometheusUrl)
	err := s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
