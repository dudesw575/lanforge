package websocket

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
)

type Stats struct {
	CPU    float64 `json:"cpu"`
	Memory uint64  `json:"memory"`
	MemMax uint64  `json:"memMax"`
}

func StreamStats(cli *client.Client, w http.ResponseWriter, r *http.Request, id string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx := context.Background()

	statsResp, err := cli.ContainerStats(ctx, id, true)
	if err != nil {
		return
	}
	defer statsResp.Body.Close()

	decoder := json.NewDecoder(statsResp.Body)

	var prevCPU float64
	var prevSystem float64

	for {
		var v map[string]interface{}
		if err := decoder.Decode(&v); err != nil {
			break
		}

		// Defensive access for cpu_stats
		cpuPercent := 0.0
		if cpuStatsRaw, ok := v["cpu_stats"].(map[string]interface{}); ok {
			cpuUsageRaw, usageOk := cpuStatsRaw["cpu_usage"].(map[string]interface{})
			systemUsage, sysOk := cpuStatsRaw["system_cpu_usage"].(float64)

			if usageOk && sysOk {
				totalUsage, totalOk := cpuUsageRaw["total_usage"].(float64)
				if totalOk {
					cpuDelta := totalUsage - prevCPU
					systemDelta := systemUsage - prevSystem

					if systemDelta > 0 && cpuDelta > 0 {
						cpuPercent = (cpuDelta / systemDelta) * 100 * float64(len(cpuUsageRaw))
					}

					prevCPU = totalUsage
					prevSystem = systemUsage
				}
			}
		}

		// Defensive access for memory_stats
		var memUsage, memLimit uint64
		if memStatsRaw, ok := v["memory_stats"].(map[string]interface{}); ok {
			if usage, ok := memStatsRaw["usage"].(float64); ok {
				memUsage = uint64(usage)
			}
			if limit, ok := memStatsRaw["limit"].(float64); ok {
				memLimit = uint64(limit)
			}
		}

		data := Stats{
			CPU:    cpuPercent,
			Memory: memUsage,
			MemMax: memLimit,
		}

		jsonData, _ := json.Marshal(data)
		if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			break
		}
	}
}
