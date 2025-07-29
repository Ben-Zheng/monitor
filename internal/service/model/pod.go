package model

import (
	"context"
	"fmt"
	"monitor/util"
)

type ModelsServer struct {
	ctx  context.Context
	from int64
	to   int64
}

func NewModelsServer(ctx context.Context, from int64, to int64) *ModelsServer {
	return &ModelsServer{
		ctx:  ctx,
		from: from,
		to:   to,
	}
}

type PodInfo struct {
	cluster     string
	clustername string
	llm_model   string
	status      string
	mem_usage   string
	cpu_usage   string
	pod         string
}

func (m *ModelsServer) GetPods(modelName string) []PodInfo {
	podsInfo := make([]PodInfo, 0)

	query := NewModelQueryReq(m.ctx, m.from, m.to)
	modelInfo, err := query.GetPodInfo("modelName", modelName)
	if err != nil {
		fmt.Println(err)
	}

	for i := range modelInfo.Matrix {
		var cpuUsage string
		var memUsage string
		var podinfo PodInfo

		pod := modelInfo.Matrix[i].Metric.Pod
		cluster := modelInfo.Matrix[i].Metric.Cluster
		clusterName := modelInfo.Matrix[i].Metric.ClusterName
		llm_model := modelInfo.Matrix[i].Metric.Llm_model
		status, err := util.SortDataPoints(modelInfo.Matrix[i].Values)
		if err != nil {
			fmt.Println(err)
		}
		cpuInfo, err := query.GetPodInfo("cpu", pod)
		if len(cpuInfo.Matrix) > 0 {
			cpuUsage, err = util.SortDataPoints(cpuInfo.Matrix[0].Values)
			if err != nil {
				fmt.Println(err)
			}
		}
		memInfo, err := query.GetPodInfo("memory", pod)
		if len(memInfo.Matrix) > 0 {
			memUsage, err = util.SortDataPoints(memInfo.Matrix[0].Values)
			if err != nil {
				fmt.Println(err)
			}

		}
		podinfo.llm_model = llm_model
		podinfo.cluster = cluster
		podinfo.clustername = clusterName
		podinfo.status = status
		podinfo.mem_usage = memUsage
		podinfo.cpu_usage = cpuUsage
		podinfo.pod = pod
		podsInfo = append(podsInfo, podinfo)
	}
	return podsInfo
}
