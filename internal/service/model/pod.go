package model

import (
	"context"
	"fmt"
	"log"
	"monitor/internal/types"
	"monitor/util"
	"net/url"
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

//type PodInfo struct {
//	cluster     string
//	clustername string
//	llm_model   string
//	status      string
//	mem_usage   string
//	cpu_usage   string
//	pod         string
//}

// 更改数据
//
//	func (m *ModelsServer) GetPods(modelName string) []PodInfo {
//		podsInfo := make([]PodInfo, 0)
//
//		query := NewModelQueryReq(m.ctx, m.from, m.to)
//		modelInfo, err := query.GetPodInfo("modelName", modelName)
//		if err != nil {
//			fmt.Println(err)
//		}
//
//		if modelInfo == nil {
//			return podsInfo
//		}
//
//		for i := range modelInfo.Matrix {
//			var cpuUsage string
//			var memUsage string
//			var podinfo PodInfo
//
//			pod := modelInfo.Matrix[i].Metric.Pod
//			cluster := modelInfo.Matrix[i].Metric.Cluster
//			clusterName := modelInfo.Matrix[i].Metric.ClusterName
//			llm_model := modelInfo.Matrix[i].Metric.Llm_model
//			status, err := util.SortDataPoints(modelInfo.Matrix[i].Values)
//			if err != nil {
//				fmt.Println(err)
//			}
//			cpuInfo, err := query.GetPodInfo("cpu", pod)
//			if len(cpuInfo.Matrix) > 0 {
//				cpuUsage, err = util.SortDataPoints(cpuInfo.Matrix[0].Values)
//				if err != nil {
//					fmt.Println(err)
//				}
//			}
//			memInfo, err := query.GetPodInfo("memory", pod)
//			if len(memInfo.Matrix) > 0 {
//				memUsage, err = util.SortDataPoints(memInfo.Matrix[0].Values)
//				if err != nil {
//					fmt.Println(err)
//				}
//			}
//			podinfo.llm_model = llm_model
//			podinfo.cluster = cluster
//			podinfo.clustername = clusterName
//			podinfo.status = status
//			podinfo.mem_usage = memUsage
//			podinfo.cpu_usage = cpuUsage
//			podinfo.pod = pod
//			podsInfo = append(podsInfo, podinfo)
//		}
//		return podsInfo
//	}
type PodInfo struct {
	url    string
	status string
}

func (m *ModelsServer) GetPods(modelName string) PodInfo {
	var podInfo PodInfo
	query := NewModelQueryReq(m.ctx, m.from, m.to)
	modelInfo, err := query.GetPodInfo("modelName", modelName)
	if err != nil {
		podInfo.status = "-"
		log.Println(err)
		return podInfo
	}

	if modelInfo == nil {
		podInfo.status = "-"
		return podInfo
	}

	sign := "success"
	if len(modelInfo.Matrix) == 0 {
		sign = "-"
	}
	gcount := 0
	bcount := 0

	for i := range modelInfo.Matrix {
		status, err := util.SortDataPoints(modelInfo.Matrix[i].Values)
		if err != nil {
			log.Println(err)
			bcount += 1
		}
		if status == "1" {
			gcount += 1
		} else if status == "0" {
			bcount += 1
		}

	}

	if gcount == 0 && bcount > 0 {
		sign = "failed"
	} else if gcount > 0 && bcount > 0 {
		sign = "unnormal"
	} else if gcount > 0 && bcount == 0 {
		sign = "success"
	}

	fullCheck_uid := types.ModelFullCheckUid
	fullCheck_path := types.ModelFullCheckPath
	url := modelDetialDashborad(fullCheck_uid, fullCheck_path, modelName)

	podInfo.status = sign
	podInfo.url = url
	return podInfo
}

func modelDetialDashborad(uid, path, modelName string) string {
	u := url.URL{
		Path: fmt.Sprintf("/ui/insight-grafana/d/%s/%s", uid, path),
	}

	q := u.Query()
	modelPrefix := "var-llm_model"
	q.Add(modelPrefix, modelName)
	qq, _ := url.ParseQuery(q.Encode())
	u.RawQuery = qq.Encode()
	return u.String()
}
