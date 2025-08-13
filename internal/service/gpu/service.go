package gpu

import (
	"context"
	"fmt"
	"log"
	"monitor/config"
	"monitor/internal/client"
	"monitor/internal/models"
	"monitor/internal/types"
	"monitor/util"
	"sync"
	"time"
)

// 定义任务类型
type task struct {
	name string
	fn   func() (map[string]int, error)
}

func GetClustersNames(ctx context.Context) []ClusterInfo {
	queryClient := NewQueryGrafana(time.Now().UnixMilli(), 0, ctx, config.GetGrafanaQueryConfig().ClusterBaseURL)
	result, err := queryClient.Getinfo("kpanda_gpu_count")
	if err != nil {
		log.Println(err)
	}
	clusterMap := make(map[string]string)
	for i := range result.Matrix {
		mode := result.Matrix[i].Metric.Mode
		cluster := result.Matrix[i].Metric.Cluster
		clusterMap[cluster] = mode
	}

	c := config.GetGrafanaQueryConfig()
	client := client.NewDCEClient(ctx, c.ClusterBaseURL, c.InsecureSkipVerify)

	cluNames := client.GetClusterName("/apis/insight.io/v1alpha1/metric/queryrange")
	clustersName := make([]ClusterInfo, 0)
	for i := range cluNames {
		clustersName = append(clustersName, ClusterInfo{
			Name:  cluNames[i].ClusterName,
			Label: cluNames[i].Cluster,
		})
	}
	return clustersName
}

func GetClustersInfo(ctx context.Context, req models.ClusterListRequest, from_timestamp int64, to_timestamp int64) *types.PagedResponse {
	baseUrl := config.GetGrafanaQueryConfig().ClusterBaseURL
	queryClient := NewQueryGrafana(from_timestamp, to_timestamp, ctx, baseUrl)
	c := config.GetGrafanaQueryConfig()
	dceclient := client.NewDCEClient(ctx, c.ClusterBaseURL, c.InsecureSkipVerify)
	cluNames := dceclient.GetClusterName("/apis/insight.io/v1alpha1/metric/queryrange")
	queryClient.SetClusterName(cluNames)

	var wg sync.WaitGroup
	type result struct {
		data map[string]int
		name string
	}
	// 任务列表
	tasks := []task{
		{"core", queryClient.GetClustersTotalCore},
		{"usedMem", queryClient.GetClustersUsedMem},
		{"totalMem", queryClient.GetClustersTotalMem},
		{"totalPvalue", queryClient.GetTotalClusterPValue},
		{"usedPvalue", queryClient.GetUsedClusterPValue},
	}

	results := make(chan result, len(tasks))
	for _, t := range tasks {
		wg.Add(1)
		go func(t task) {
			defer wg.Done()
			data, err := t.fn()
			if err != nil {
				log.Println(err)
				data = make(map[string]int)
			}
			results <- result{data: data, name: t.name}
		}(t)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	dataMap := make(map[string]map[string]int)
	for res := range results {
		for label, val := range res.data {
			if dataMap[label] == nil {
				dataMap[label] = make(map[string]int)
			}
			dataMap[label][res.name] = val
		}
	}

	clusters := make([]models.Cluster, 0, len(cluNames))
	for _, cn := range cluNames {
		metrics := dataMap[cn.Cluster]
		clusters = append(clusters, models.Cluster{
			Name:           cn.ClusterName,
			Label:          cn.Cluster,
			GPUCores:       metrics["core"],
			UsedGPUMemory:  metrics["usedMem"],
			TotalGPUMemory: metrics["totalMem"],
			UsedPValue:     float64(metrics["usedPvalue"]) / 100.0,
			TotalPValue:    float64(metrics["totalPvalue"]) / 100.0,
		})
	}
	if req.Size <= 0 {
		req.Size = 10 // 设置默认页大小
	}
	totalItems := len(clusters)
	totalPages := totalItems / req.Size
	if totalItems%req.Size != 0 {
		totalPages++
	}

	start := (req.Page - 1) * req.Size
	if start < 0 {
		start = 0
	}
	end := start + req.Size
	if end > totalItems {
		end = totalItems
	}

	hasNext := req.Page < totalPages

	return &types.PagedResponse{
		Page:       req.Page,
		PageSize:   req.Size,
		TotalPages: totalPages,
		TotalItems: totalItems,
		HasNext:    hasNext,
		Data:       clusters[start:end],
	}
}

func GetNodesInfo(ctx context.Context, req models.NodesListRequest, from_timestamp int64, to_timestamp int64, baseUrl string) *types.PagedNodesResponse {
	queryClient := NewQueryGrafana(from_timestamp, to_timestamp, ctx, baseUrl)
	nodesName, _ := queryClient.GetNodesName(req.Cluster)
	queryClient.SetClusterId(req.Cluster)
	queryClient.SetModeStr(req.Cluster)
	// 任务列表
	tasks := []task{
		{"core", queryClient.GetNodeTotal},
		{"usedMem", queryClient.GetNodesUsedMem},
		{"totalMem", queryClient.GetNodesTotalMem},
		{"totalPvalue", queryClient.GetNodeTotalPValue},
		{"usedPvalue", queryClient.GetNodeUsedPValue},
	}
	var wg sync.WaitGroup
	type result struct {
		data map[string]int
		name string
	}
	results := make(chan result, len(tasks))
	for _, t := range tasks {
		wg.Add(1)
		go func(t task) {
			defer wg.Done()
			data, err := t.fn()
			if err != nil {
				log.Println(err)
				data = make(map[string]int)
			}
			results <- result{data: data, name: t.name}
		}(t)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	dataMap := make(map[string]map[string]int)
	for res := range results {
		for label, val := range res.data {
			if dataMap[label] == nil {
				dataMap[label] = make(map[string]int)
			}
			dataMap[label][res.name] = val
		}
	}

	nodes := make([]models.Node, 0, len(nodesName))
	for _, cn := range nodesName {
		metrics := dataMap[cn.Name]
		nodes = append(nodes, models.Node{
			Name:           cn.Name,
			GPUCores:       metrics["core"],
			UsedGPUMemory:  metrics["usedMem"],
			TotalGPUMemory: metrics["totalMem"],
			PvalueTotal:    float64(metrics["usedPvalue"]) / 100.0,
			PvalueUsed:     float64(metrics["totalPvalue"]) / 100.0,
		})

	}

	totalItems := len(nodes) // 总记录数
	totalPages := totalItems / req.Size
	if totalItems%req.Size != 0 {
		totalPages++
	}

	start := (req.Page - 1) * req.Size
	if start < 0 {
		start = 0
	}
	end := start + req.Size
	if end > totalItems {
		end = totalItems
	}

	hasNext := req.Page < totalPages
	// 返回分页响应
	return &types.PagedNodesResponse{
		Page:       req.Page,
		PageSize:   req.Size,
		TotalPages: totalPages,
		TotalItems: totalItems,
		HasNext:    hasNext,
		Data:       nodes[start:end],
	}
}

func GetClusterDetailByModel(ctx context.Context, param models.DetailRequest, from_timestamp int64, to_timestamp int64, baseUrl string, modeStr string) []models.ModelDetailResponse {
	queryClient := NewQueryGrafana(from_timestamp, to_timestamp, ctx, baseUrl)
	res, err := queryClient.GetClusterUsedDetailByModel(param.Cluster, modeStr)
	if err != nil {
		log.Println(err)
	}
	resTotal, err := queryClient.GetClusterTotalDetailByModel(param.Cluster, modeStr)
	if err != nil {
		log.Println(err)
	}
	//当前节点的分型号缓存和当前集群的分型号缓存
	var total int
	for _, v := range res {
		total += v
	}
	modeldetailResponse := make([]models.ModelDetailResponse, 0)

	for k, v := range res {
		resp := models.ModelDetailResponse{
			NodeName:    k,
			UsedMem:     v,
			UsedPercent: util.SafePercent(v, total),
			TotalMem:    resTotal[k],
		}
		modeldetailResponse = append(modeldetailResponse, resp)
	}
	return modeldetailResponse
}

func GetClusterDetailPValueByModel(ctx context.Context, param models.DetailRequest, from_timestamp, to_timestamp int64, baseUrl string, modStr string, cluNames []types.NameList) []models.ModelPvalueResponse {
	queryClient := NewQueryGrafana(from_timestamp, to_timestamp, ctx, baseUrl)
	queryClient.SetClusterName(cluNames)
	pvalueMap, err := queryClient.CalClusterDetailPValueByModel(param.Cluster, modStr)
	if err != nil {
		fmt.Println(err)
	}
	Pvalues := make([]models.ModelPvalueResponse, 0)
	for k, v := range pvalueMap {
		var p models.ModelPvalueResponse
		p.ModelName = k
		val := float64(v) / 100.0
		p.Pvalue = val
		Pvalues = append(Pvalues, p)
	}
	return Pvalues

}

func GetNodesPvalueDetailByModel(ctx context.Context, param models.DetailRequest, from_timestamp, to_timestamp int64, baseUrl string, modeStr string) []models.ModelPvalueResponse {
	queryClient := NewQueryGrafana(from_timestamp, to_timestamp, ctx, baseUrl)
	c := config.GetGrafanaQueryConfig()
	client := client.NewDCEClient(ctx, c.ClusterBaseURL, c.InsecureSkipVerify)
	cluNames := client.GetClusterName("/apis/insight.io/v1alpha1/metric/queryrange")
	queryClient.SetClusterName(cluNames)
	pvalueMap, err := queryClient.CalNodesPvalueDetailByModel(param.Cluster, param.Node, modeStr)

	if err != nil {
		fmt.Println(err)
	}
	Pvalues := make([]models.ModelPvalueResponse, 0)
	for k, v := range pvalueMap {
		var p models.ModelPvalueResponse
		p.ModelName = k
		val := float64(v) / 100.0
		p.Pvalue = val
		Pvalues = append(Pvalues, p)
	}

	return Pvalues
}

func GetNodesDetailByModel(ctx context.Context, param models.DetailRequest, from_timestamp, to_timestamp int64, baseUrl string) models.NodeDetailResponse {
	queryClient := NewQueryGrafana(from_timestamp, to_timestamp, ctx, baseUrl)
	c := config.GetGrafanaQueryConfig()
	client := client.NewDCEClient(ctx, c.ClusterBaseURL, c.InsecureSkipVerify)
	cluNames := client.GetClusterName("/apis/insight.io/v1alpha1/metric/queryrange")
	queryClient.SetClusterName(cluNames)

	res, err := queryClient.GetNodeUsedDetailByNode(param.Cluster, param.Node)
	if err != nil {
		fmt.Println(err)
	}
	totalMem, err := queryClient.GetNodeTotalDetailByNode(param.Cluster, param.Node)
	if err != nil {
		fmt.Println(err)
	}
	//当前节点的分型号缓存和当前集群的分型号缓存
	total := totalMem[param.Node]

	return models.NodeDetailResponse{
		NodeName:    param.Node,
		UsedMem:     res[param.Node],
		TotalMem:    totalMem[param.Node],
		UsedPercent: util.SafePercent(res[param.Node], total),
	}
}

func GetCoresByCluster(ctx context.Context, param models.DetailRequest, from_timestamp, to_timestamp int64, baseUrl string, cluNames []types.NameList) models.CoreResponse {
	queryClient := NewQueryGrafana(from_timestamp, to_timestamp, ctx, baseUrl)
	queryClient.SetClusterName(cluNames)

	totalCores, err := queryClient.GetClustersTotalCore()
	if err != nil {
		fmt.Println(err)
	}
	usedCores, err := queryClient.GetClustersUsedCore()
	if err != nil {
		fmt.Println(err)
	}
	return models.CoreResponse{
		TotalCore: totalCores[param.Cluster],
		UsedCore:  usedCores[param.Cluster],
	}
}

func GetCoresByNode(ctx context.Context, param models.DetailRequest, from_timestamp, to_timestamp int64, baseUrl string) models.CoreResponse {
	queryClient := NewQueryGrafana(from_timestamp, to_timestamp, ctx, baseUrl)

	totalCores, err := queryClient.GetNodeTotalCore(param.Cluster, param.Node)
	if err != nil {
		fmt.Println(err)
	}
	usedCores, err := queryClient.GetNodeUsedCore(param.Cluster, param.Node)
	if err != nil {
		fmt.Println(err)
	}
	return models.CoreResponse{
		TotalCore: totalCores[param.Node],
		UsedCore:  usedCores[param.Node],
	}
}
