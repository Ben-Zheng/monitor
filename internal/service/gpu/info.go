package gpu

import (
	"context"
	"fmt"
	"monitor/config"
	"strings"
)

type QueryExpr struct{}

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
