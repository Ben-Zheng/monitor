package gpu

import (
	"context"
	"fmt"
	"log"
	"monitor/config"
	"monitor/internal/types"
	"monitor/util"
	"strings"
)

type PvalueDetailResp struct {
	NodesNum    int
	Pvalue      float64
	TotalPvalue float64
	Cards       int
}

type QueryExpr struct {
}

func NewQueryExpr() *QueryExpr {
	return &QueryExpr{}
}
func GetClusterType(ctx context.Context, clusterId string, fromstamp, tostamp int64) string {
	client := NewQueryGrafana(fromstamp, tostamp, ctx, config.GetGrafanaQueryConfig().ClusterBaseURL)
	result, err := client.Getinfo("kpanda_gpu_count")
	if err != nil {
		fmt.Println(err)
	}
	clusterMap := make(map[string]string)
	for i := range result.Matrix {
		mode := result.Matrix[i].Metric.Mode
		cluster := result.Matrix[i].Metric.Cluster
		clusterMap[cluster] = mode
	}
	return clusterMap[clusterId]
}

func (q *QueryExpr) GetTotalNvidiaPvalue(ctx context.Context, from, to int64) (map[string]*PvalueDetailResp, error) {
	client := NewQueryGrafana(from, to, ctx, config.GetGrafanaQueryConfig().ClusterBaseURL)
	expr := q.makeExprTotalCoreGPU()
	result, err := client.Getinfo(expr)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	if result == nil || len(result.Matrix) == 0 {
		return nil, err
	}

	infoMap := make(map[string]*PvalueDetailResp)
	var totalPvalue float64
	for i := range result.Matrix {
		data := util.ExtractValues(result.Matrix[i].Values)
		avgData, _ := util.CalculateMedian(data)
		label := result.Matrix[i].Metric.ModelName
		ruleMap := client.GetRule()

		if _, exists := infoMap[label]; !exists {
			infoMap[label] = &PvalueDetailResp{}
		}

		addPvalue := float64(avgData) * ruleMap[label]
		infoMap[label].Pvalue += addPvalue
		totalPvalue += addPvalue
		infoMap[label].NodesNum++
		infoMap[label].Cards += avgData
	}
	return infoMap, nil
}

func (q *QueryExpr) GetTotalAscendPvalue(ctx context.Context, from, to int64) (map[string]*PvalueDetailResp, error) {
	client := NewQueryGrafana(from, to, ctx, config.GetGrafanaQueryConfig().ClusterBaseURL)
	expr := q.makeExprTotalCoreNPU()
	result, err := client.Getinfo(expr)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	if result == nil || len(result.Matrix) == 0 {
		return nil, err
	}
	infoMap := make(map[string]*PvalueDetailResp)
	var totalPvalue float64
	for i := range result.Matrix {
		data := util.ExtractValues(result.Matrix[i].Values)
		avgData, _ := util.CalculateMedian(data)
		label := result.Matrix[i].Metric.Model_Name
		ruleMap := client.GetRule()

		if _, exists := infoMap[label]; !exists {
			infoMap[label] = &PvalueDetailResp{}
		}

		addPvalue := float64(avgData) * ruleMap[label]
		infoMap[label].Pvalue += addPvalue
		totalPvalue += addPvalue
		infoMap[label].NodesNum++
		infoMap[label].Cards += avgData
	}
	return infoMap, nil
}

func (q *QueryExpr) Getmodel_node_view(ctx context.Context, fromstamp, tostamp int64) ([]types.ModelNodeViewDetails, error) {
	client := NewQueryGrafana(fromstamp, tostamp, ctx, config.GetGrafanaQueryConfig().ClusterBaseURL)
	expr := q.makeExprchip_info_model_core_utilization()
	result, err := client.Getinfo(expr)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	modelViewDetails := make([]types.ModelNodeViewDetails, 0)
	for i := range result.Matrix {
		var model_view_detail types.ModelNodeViewDetails
		node := result.Matrix[i].Metric.Node
		resource := result.Matrix[i].Metric.Resource
		modelName := result.Matrix[i].Metric.Label_llm_model
		data := util.ExtractValues(result.Matrix[i].Values)
		mediaData, _ := util.CalculateMedian(data)

		model_view_detail.Label_llm_model = modelName
		model_view_detail.Node = node
		model_view_detail.Resource = resource
		model_view_detail.Core = mediaData

		ruleMap := client.GetRule()
		pvalue := ruleMap[resource] * float64(mediaData)
		model_view_detail.Pvalue = pvalue
		modelViewDetails = append(modelViewDetails, model_view_detail)
	}
	return modelViewDetails, nil
}

func (q *QueryExpr) makeTotalExpr(modelName, info, level, kind, clusterId, nodeId string) string {
	var model string
	var expr string
	if strings.Contains(modelName, "Nvidia") {
		model = "DCGM_FI_DEV_FB_TOTAL"
	} else if strings.Contains(modelName, "Ascend") {
		model = "npu_chip_info_hbm_total_memory"
		info = strings.ReplaceAll(info, "modelName", "model_name")
	}
	if level == "all" {
		expr = fmt.Sprintf("%s by (%s) (%s{cluster!=\"\"})", kind, info, model)
	} else if level == "cluster" {
		expr = fmt.Sprintf("%s by (%s) (%s{cluster=\"%s\"})", kind, info, model, clusterId)
	} else if level == "node" {
		expr = fmt.Sprintf("%s by (%s) (%s{cluster=\"%s\",node=\"%s\"})", kind, info, model, clusterId, nodeId)
	} else if level == "allnode" {
		expr = fmt.Sprintf("%s by (%s) (%s{cluster=\"%s\",node!=\"\"})", kind, info, model, clusterId)
	}
	return expr
}

func (q *QueryExpr) makeUsedExpr(modelName, info, level, kind, clusterId, nodeId string) string {
	var model string
	var expr string
	if strings.Contains(modelName, "Nvidia") {
		model = "DCGM_FI_DEV_FB_USED"
	} else if strings.Contains(modelName, "Ascend") {
		model = "npu_chip_info_hbm_used_memory"
		info = strings.ReplaceAll(info, "modelName", "model_name")
	}

	if level == "all" {
		expr = fmt.Sprintf("%s by (%s) (%s{cluster!=\"\"})", kind, info, model)

	} else if level == "cluster" {
		expr = fmt.Sprintf("%s by (%s) (%s{cluster=\"%s\"})", kind, info, model, clusterId)

	} else if level == "node" {
		expr = fmt.Sprintf("%s by (%s) (%s{cluster=\"%s\",node=\"%s\"})", kind, info, model, clusterId, nodeId)

	} else if level == "allnode" {
		expr = fmt.Sprintf("%s by (%s) (%s{cluster=\"%s\",node!=\"\"})", kind, info, model, clusterId)
	}
	return expr
}

func (q *QueryExpr) makeExprKpandaCluster(modelName string, clusterId string) string {
	var expr string
	if modelName == "clusters" {
		expr = fmt.Sprintf("kpanda_gpu_allocated")
	} else if modelName == "cluster" {
		expr = fmt.Sprintf("kpanda_gpu_allocated{cluster=\"%s\"}", clusterId)
	}
	return expr
}

func (q *QueryExpr) makeExprKpandaNode(clusterId string, node string) string {
	expr := fmt.Sprintf("kpanda_gpu_allocated{cluster=\"%s\",node=\"%s\"}", clusterId, node)
	return expr
}

func (q *QueryExpr) makeExprKpandainfo_utilization(clusterId string, node string) string {
	expr := fmt.Sprintf("kpanda_gpu_pod_utilization{cluster=\"%s\",node=\"%s\"}", clusterId, node)
	return expr
}

func (q *QueryExpr) makeExprchip_info_utilization(clusterId string, node string) string {
	expr := fmt.Sprintf("npu_chip_info_utilization{cluster=\"%s\",node=\"%s\"}", clusterId, node)
	return expr
}

func (q *QueryExpr) makeExprchip_info_model_core_utilization() string {
	expr := fmt.Sprintf("kube_pod_container_resource_requests{resource=~\"huawei_com_Ascend.*|nvidia_com_gpu.*\"} " +
		" * on(namespace, pod) group_left(label_llm_model, node, host_ip) (kube_pod_info * on(namespace, pod) group_left(label_llm_model) kube_pod_labels{label_llm_model!=\"\"})")
	return expr
}

func (q *QueryExpr) makeExprTotalCoreGPU() string {
	expr := fmt.Sprintf("count by (modelName)(DCGM_FI_DEV_FB_TOTAL)")
	return expr
}

func (q *QueryExpr) makeExprTotalCoreNPU() string {
	expr := fmt.Sprintf("count by (model_name)(npu_chip_info_hbm_total_memory)")
	return expr
}
