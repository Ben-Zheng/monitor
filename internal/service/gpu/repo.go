package gpu

import "monitor/internal/types"

type QueryGrafanaInfoRepo interface {
	GetClustersMemMap(nvidiaMem, ascendMem map[string]int) (map[string]int, error)

	GetClustersUsedMem() (map[string]int, error)
	GetClustersTotalMem() (map[string]int, error)

	GetNodesUsedMem() (map[string]int, error)
	GetNodesTotalMem() (map[string]int, error)

	GetNodeTotalCore(clusterID, nodeId string) (map[string]int, error)
	GetNodeUsedCore(clusterID, nodeId string) (map[string]int, error)

	GetClustersTotalCore() (map[string]int, error)
	GetClustersUsedCore() (map[string]int, error)
	GetClusterTotalDetailByModel(clusterId string, modeStr string) (map[string]int, error)
	GetClusterUsedDetailByModel(clusterId string, modeStr string) (map[string]int, error)

	GetNodeTotalDetailByNode(clusterId string, NodeId string) (map[string]int, error)
	GetNodeUsedDetailByNode(clusterId string, NodeId string) (map[string]int, error)
	GetNodeTotal() (map[string]int, error)
	CalNodesPvalueDetailByModel(clusterId, nodeId, modeStr string) (map[string]int, error)

	GetNodeTotalPValue() (map[string]int, error)
	GetNodeUsedPValue() (map[string]int, error)

	GetTotalClusterPValue() (map[string]int, error)
	GetUsedClusterPValue() (map[string]int, error)
	CalClusterDetailPValueByModel(clusterId string, modeStr string) (map[string]int, error)

	GetNodesName(clusterID string) ([]NodeInfo, error)
	Getinforange(expr string) (*types.VectorResponse, error)
	Getinfo(expr string) (*types.VectorResponse, error)

	SetClusterName(names []types.NameList)
	SetClusterId(clusterId string)
	SetNodeId(nodeId string)
	GetModeStr(clusterId string) string
	SetModeStr(clusterId string)
	GetRule() map[string]float64
}
