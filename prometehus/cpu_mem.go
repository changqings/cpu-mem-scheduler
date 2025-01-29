package prometehus

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	p8sv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type CpuMem [2]float64

func GetNodeMetrics(ctx context.Context, client api.Client) (map[string]CpuMem, error) {

	v1api := p8sv1.NewAPI(client)

	timeNow := time.Now()

	cpuQuery := `sum(irate(container_cpu_usage_seconds_total{image!=""}[5m]))by(node)`
	cpuRes, _, err := v1api.Query(ctx, cpuQuery, timeNow, p8sv1.WithTimeout(10*time.Second))
	if err != nil {
		return nil, err
	}

	memQuery := `1-node_memory_MemAvailable_bytes/node_memory_MemTotal_bytes`
	memRes, _, err := v1api.Query(ctx, memQuery, timeNow, p8sv1.WithTimeout(10*time.Second))
	if err != nil {
		return nil, err
	}

	cpuUsageVecSample := cpuRes.(model.Vector)
	memUsageVecSample := memRes.(model.Vector)

	nodeCpuMap := make(map[string]model.SampleValue)
	for _, sample := range cpuUsageVecSample {
		nodeName := string(sample.Metric["node"])
		nodeCpuMap[nodeName] = sample.Value
	}

	nodeMemMap := make(map[string]model.SampleValue)
	for _, sample := range memUsageVecSample {

		instance := string(sample.Metric["instance"])
		nodeName, ok := strings.CutSuffix(instance, ":9100")
		if ok {
			nodeMemMap[nodeName] = sample.Value
		}
	}

	cpuMemMap := make(map[string]CpuMem)

	for nodeName, cpuUsage := range nodeCpuMap {
		memUsage, ok := nodeMemMap[nodeName]
		if !ok {
			continue
		}
		cpuMemMap[nodeName] = CpuMem{float64(cpuUsage), float64(memUsage)}
	}

	return cpuMemMap, nil
}

func GetP8sClient(url string) (api.Client, error) {
	cfg := api.Config{
		Address: url,
	}
	return api.NewClient(cfg)
}
