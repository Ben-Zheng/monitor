package model

import (
	"fmt"
)

func BuildModelQueryRange(data_name, keyword, start, end string) map[string]string {
	query := make(map[string]string, 0)
	var expr string
	if data_name == "" {
		return query
	}
	if data_name == "cpu" {
		expr = makeCPUExpr(keyword)
	}
	if data_name == "memory" {
		expr = makeMemExpr(keyword)
	}
	if data_name == "modelName" {
		expr = makeProbeExpr(keyword)
	}

	return map[string]string{
		"query": expr,
		"start": start,
		"end":   end,
	}
}

func makeProbeExpr(modelName string) string {
	var expr string
	expr = fmt.Sprintf("probe_success{llm_model=\"%s\"}", modelName)
	return expr
}

func makeCPUExpr(pod string) string {
	var expr string
	expr = fmt.Sprintf("pod_cpu_utilization{pod=\"%s\"}", pod)
	return expr
}

func makeMemExpr(pod string) string {
	var expr string
	expr = fmt.Sprintf("pod_memory_utilization{pod=\"%s\"}", pod)
	return expr
}
