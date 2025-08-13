package gpu

import (
	"context"
	"fmt"
	"monitor/config"
	"monitor/internal/client"
	"monitor/internal/types"
	"strconv"
	"strings"
)

func NewQueryGrafana(from, to int64, ctx context.Context, baseUrl string) QueryGrafanaInfoRepo {
	headers := map[string]string{
		"content-type":    "application/json",
		"Accept":          "application/json, text/plain, */*",
		"Accept-Encoding": "gzip, deflate",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
	}
	modelFP := config.GetFPRule()
	return &QueryGrafana{
		Url:    baseUrl,
		From:   from,
		To:     to,
		Header: headers,
		Rule:   modelFP,
		Ctx:    ctx,
	}
}

type QueryGrafana struct {
	Ctx         context.Context
	Url         string             `json:"url"`
	From        int64              `json:"from"`
	To          int64              `json:"to"`
	Header      map[string]string  `json:"header"`
	Rule        map[string]float64 `json:"rule"`
	ClusterName []types.NameList   `json:"cluster_name"`

	ClusterId string `json:"cluster_id"`
	NodeId    string `json:"node_id"`
	ModeStr   string `json:"mode_str"`
}

type ClusterInfo struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type NodeInfo struct {
	Name string `json:"name"`
}

func (q *QueryGrafana) SetClusterId(clusterId string) {
	q.ClusterId = clusterId
}
func (q *QueryGrafana) SetNodeId(nodeId string) {
	q.NodeId = nodeId
}
func (q *QueryGrafana) GetModeStr(clusterId string) string {
	modeStr := GetClusterType(q.Ctx, clusterId, q.From, q.To)
	return modeStr
}

func (q *QueryGrafana) GetClustersMemMap(nvidiaMem, ascendMem map[string]int) (map[string]int, error) {
	names := q.ClusterName
	usedMems := make(map[string]int, len(names))
	for i := range names {
		usedMem := nvidiaMem[names[i].Cluster] + ascendMem[names[i].Cluster]
		usedMems[names[i].Cluster] = usedMem
	}
	return usedMems, nil
}

func (q *QueryGrafana) GetClustersUsedMem() (map[string]int, error) {
	handler := NewClusterHandler(q)
	nvidiaMem, _ := handler.getUsedMemByClusterAll("Nvidia")
	ascendMem, _ := handler.getUsedMemByClusterAll("Ascend")
	mems, err := q.GetClustersMemMap(nvidiaMem, ascendMem)
	return mems, err
}

func (q *QueryGrafana) GetClustersTotalMem() (map[string]int, error) {
	handler := NewClusterHandler(q)
	nvidiaMem, _ := handler.getTotalMemByClusterAll("Nvidia")
	ascendMem, _ := handler.getTotalMemByClusterAll("Ascend")
	mems, err := q.GetClustersMemMap(nvidiaMem, ascendMem)
	return mems, err
}

func (q *QueryGrafana) GetNodesUsedMem() (map[string]int, error) {
	clusterID := q.ClusterId
	handler := NewHandlerNode(q, clusterID)
	nodeMem := make(map[string]int)
	if strings.ToLower(handler.modeStr) == "gpu" {
		nodeMem, _ = handler.getUsedMemByNodeAll("Nvidia", clusterID)
	} else if strings.ToLower(handler.modeStr) == "npu" {
		nodeMem, _ = handler.getUsedMemByNodeAll("Ascend", clusterID)
	}

	return nodeMem, nil
}

func (q *QueryGrafana) GetNodesTotalMem() (map[string]int, error) {
	clusterID := q.ClusterId
	handler := NewHandlerNode(q, clusterID)
	nodeMem := make(map[string]int)
	if strings.ToLower(handler.modeStr) == "gpu" {
		nodeMem, _ = handler.getTotalMemByNodeAll("Nvidia", clusterID)
	} else if strings.ToLower(handler.modeStr) == "npu" {
		nodeMem, _ = handler.getTotalMemByNodeAll("Ascend", clusterID)
	}
	return nodeMem, nil
}

func (q *QueryGrafana) GetNodeTotalCore(clusterID, nodeId string) (map[string]int, error) {
	handler := NewHandlerNode(q, clusterID)
	nodeMem := make(map[string]int)
	if strings.ToLower(handler.modeStr) == "gpu" {
		nodeMem, _ = handler.getTotalCoreByNode("Nvidia", clusterID, nodeId)
	} else if strings.ToLower(handler.modeStr) == "npu" {
		nodeMem, _ = handler.getTotalCoreByNode("Ascend", clusterID, nodeId)
	}
	return nodeMem, nil
}

func (q *QueryGrafana) GetNodeUsedCore(clusterID, nodeId string) (map[string]int, error) {
	handler := NewHandlerNode(q, clusterID)
	nodeMem := make(map[string]int)
	if strings.ToLower(handler.modeStr) == "gpu" {
		nodeMem, _ = handler.getUsedCoreByNode("Nvidia", clusterID, nodeId)
	} else if strings.ToLower(handler.modeStr) == "npu" {
		nodeMem, _ = handler.getUsedCoreByNode("Ascend", clusterID, nodeId)
	}
	return nodeMem, nil
}

func (q *QueryGrafana) GetClustersTotalCore() (map[string]int, error) {

	handler := NewHandlerCore(q)

	nvidiaCore, _ := handler.getNvidiaTotalCore()
	ascendCore, _ := handler.getAscendTotalCore()
	names := handler.query.ClusterName
	totalCores := make(map[string]int, len(names))
	for i := range names {
		totalCore := nvidiaCore[names[i].Cluster] + ascendCore[names[i].Cluster]
		totalCores[names[i].Cluster] = totalCore
	}
	return totalCores, nil
}

func (q *QueryGrafana) GetClustersUsedCore() (map[string]int, error) {
	handler := NewHandlerCore(q)
	totalCores, err := handler.getKpandaGpu()
	return totalCores, err
}

func (q *QueryGrafana) GetNodeTotalDetailByNode(clusterId string, NodeId string) (map[string]int, error) {
	handler := NewHandlerNode(q, clusterId)
	nodeMem := make(map[string]int)
	if strings.ToLower(handler.modeStr) == "gpu" {
		nodeMem, _ = handler.getTotalMemByNode("Nvidia", clusterId, NodeId)
	} else if strings.ToLower(handler.modeStr) == "npu" {
		nodeMem, _ = handler.getTotalMemByNode("Ascend", clusterId, NodeId)
	}
	return nodeMem, nil
}

func (q *QueryGrafana) GetNodeUsedDetailByNode(clusterId string, NodeId string) (map[string]int, error) {
	handler := NewHandlerNode(q, clusterId)
	nodeMem := make(map[string]int)
	if strings.ToLower(handler.modeStr) == "gpu" {
		nodeMem, _ = handler.getUsedMemByNode("Nvidia", clusterId, NodeId)
	} else if strings.ToLower(handler.modeStr) == "npu" {
		nodeMem, _ = handler.getUsedMemByNode("Ascend", clusterId, NodeId)
	}
	return nodeMem, nil
}
func (q *QueryGrafana) CalNodesPvalueDetailByModel(clusterId, nodeId, modeStr string) (nodeMem map[string]int, err error) {
	handler := NewHandlerNode(q, clusterId)
	nodeMem = make(map[string]int)
	if strings.ToLower(modeStr) == "gpu" {
		nodeMem, _ = handler.getUsedPvalueByNode("Nvidia", clusterId, nodeId)
	} else if strings.ToLower(modeStr) == "npu" {
		nodeMem, _ = handler.getUsedPvalueByNode("Ascend", clusterId, nodeId)
	}
	return nodeMem, nil
}

func (q *QueryGrafana) GetNodesName(clusterID string) ([]NodeInfo, error) {
	handler := NewHandlerNode(q, clusterID)
	nodes := handler.getNodesName(clusterID)
	return nodes, nil
}

func (q *QueryGrafana) GetNodeTotalPValue() (map[string]int, error) {
	clusterId := q.ClusterId
	handler := NewHandlerPvalue(q)
	totalCore := make(map[string]int)
	if strings.ToLower(handler.modeStr) == "gpu" {
		totalCore = handler.getCountTotalByModelNameAllNode("Nvidia", clusterId)
	} else if strings.ToLower(handler.modeStr) == "npu" {
		totalCore = handler.getCountTotalByModelNameAllNode("Ascend", clusterId)
	}
	return totalCore, nil
}

func (q *QueryGrafana) GetNodeUsedPValue() (map[string]int, error) {
	clusterId := q.ClusterId
	handler := NewHandlerPvalue(q)
	usedCore := make(map[string]int)
	if strings.ToLower(handler.modeStr) == "npu" {
		usedCore = handler.getCountUsedByModelNameAllNode("Ascend", clusterId)
	} else if strings.ToLower(handler.modeStr) == "gpu" {
		usedCore = handler.getCountUsedByModelNameAllNode("Nvidia", clusterId)
	}
	return usedCore, nil
}

func (q *QueryGrafana) GetNodeTotal() (map[string]int, error) {
	clusterID := q.ClusterId
	coreHandler := NewHandlerCore(q)
	totalCore := make(map[string]int)
	if strings.ToLower(coreHandler.modeStr) == "gpu" {
		totalCore, _ = coreHandler.getNvidiaNodesTotalMCore(clusterID)
	} else if strings.ToLower(coreHandler.modeStr) == "npu" {
		totalCore, _ = coreHandler.getAscendNodesTotalMCore(clusterID)
	}

	return totalCore, nil
}

func (q *QueryGrafana) GetClusterTotalDetailByModel(clusterId string, modeStr string) (map[string]int, error) {
	handler := NewHandlerModelName(q)
	GpuMap := make(map[string]int)
	var err error
	if strings.ToLower(modeStr) == "gpu" {
		GpuMap, err = handler.getTotalMemByModelNameCluster("Nvidia", clusterId)
	} else if strings.ToLower(modeStr) == "npu" {
		GpuMap, err = handler.getTotalMemByModelNameCluster("Ascend", clusterId)
	}

	return GpuMap, err
}
func (q *QueryGrafana) GetClusterUsedDetailByModel(clusterId string, modeStr string) (map[string]int, error) {
	handler := NewHandlerModelName(q)
	GpuMap := make(map[string]int)
	var err error
	if strings.ToLower(modeStr) == "gpu" {
		GpuMap, err = handler.getUsedMemByModelNameCluster("Nvidia", clusterId)
	} else if strings.ToLower(modeStr) == "npu" {
		GpuMap, err = handler.getUsedMemByModelNameCluster("Ascend", clusterId)
	}

	return GpuMap, err
}
func (q *QueryGrafana) GetTotalClusterPValue() (map[string]int, error) {
	names := q.ClusterName
	mapTotalPvalue := make(map[string]int)
	handler := NewHandlerPvalue(q)

	for i := range names {
		cluster := names[i].Cluster
		TotalPvalue := handler.getClusterTotalPValue(cluster)
		mapTotalPvalue[cluster] = TotalPvalue
	}
	return mapTotalPvalue, nil
}

func (q *QueryGrafana) GetUsedClusterPValue() (map[string]int, error) {
	names := q.ClusterName
	handler := NewHandlerPvalue(q)
	mapUsedPvalue := make(map[string]int)
	var err error
	for i := range names {
		cluster := names[i].Cluster
		coreNum := handler.getCountUsedcore("cluster", cluster)
		totalCores, err := q.GetClustersTotalCore()
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		TotalPvalue := handler.getClusterTotalPValue(cluster)
		UsedPvalue := float64(TotalPvalue) * (float64(coreNum[cluster]) / float64(totalCores[cluster]))
		mapUsedPvalue[cluster] = int(UsedPvalue)
	}
	return mapUsedPvalue, err
}

func (q *QueryGrafana) CalClusterDetailPValueByModel(clusterId string, modeStr string) (map[string]int, error) {
	handler := NewHandlerPvalue(q)
	Cores := make(map[string]int)
	if strings.ToLower(modeStr) == "gpu" {
		Cores = handler.getCountTotalByModelNameCluster("Nvidia", clusterId)
	} else if strings.ToLower(modeStr) == "npu" {
		Cores = handler.getCountTotalByModelNameCluster("Ascend", clusterId)
	}
	Pvalue := calPvalue(q.Rule, Cores)
	return Pvalue, nil
}

func (q *QueryGrafana) Getinfo(expr string) (*types.VectorResponse, error) {
	query := config.NewQueryConfig()
	queryStr := query.BuildGrafanaQueryRange(expr, strconv.FormatInt(q.From, 10), strconv.FormatInt(q.To, 10))
	client := client.NewDCEClient(q.Ctx, q.Url, query.InsecureSkipVerify)
	result, err := client.MakeGetReqRange(queryStr, "/apis/insight.io/v1alpha1/metric/queryrange")
	if err != nil {
		fmt.Println(err)
	}

	return result, err
}
func (q *QueryGrafana) Getinforange(expr string) (*types.VectorResponse, error) {
	query := config.NewQueryConfig()
	f := strconv.Itoa(int(q.From))
	t := strconv.Itoa(int(q.To))
	queryRange := query.BuildGrafanaQueryRange(expr, f, t)
	client := client.NewDCEClient(q.Ctx, q.Url, query.InsecureSkipVerify)
	result, err := client.MakeGetReqRange(queryRange, "/apis/insight.io/v1alpha1/metric/queryrange")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return result, err
}

func (q *QueryGrafana) SetClusterName(names []types.NameList) {
	q.ClusterName = names
}

func (q *QueryGrafana) SetModeStr(clusterId string) {
	modeStr := GetClusterType(q.Ctx, clusterId, q.From, q.To)
	q.ModeStr = modeStr
}

func (q *QueryGrafana) GetRule() map[string]float64 {
	return q.Rule
}
