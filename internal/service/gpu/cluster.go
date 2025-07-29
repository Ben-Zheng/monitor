package gpu

import (
	"fmt"
	"monitor/internal/types"
	"monitor/util"
)

type ClusterHandler struct {
	query     *QueryGrafana
	modelName string
}

func NewClusterHandler(query *QueryGrafana) *ClusterHandler {
	return &ClusterHandler{
		query: query,
	}
}

// 所有集群的已用显存和所有显存
func (q *ClusterHandler) getUsedMemByClusterAll(modelName string) (map[string]int, error) {
	queryStr := NewQueryExpr()
	expr := queryStr.makeUsedExpr(modelName, "cluster", "all", "sum", "", "")
	result, err := q.query.Getinforange(expr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	usedMem := getMapInfoCluster(result)
	return usedMem, nil
}
func (q *ClusterHandler) getTotalMemByClusterAll(modelName string) (map[string]int, error) {
	queryStr := NewQueryExpr()
	expr := queryStr.makeTotalExpr(modelName, "cluster", "all", "sum", "", "")
	result, err := q.query.Getinforange(expr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	usedMem := getMapInfoCluster(result)
	return usedMem, nil
}

func getMapInfoCluster(result *types.VectorResponse) map[string]int {
	infoMap := make(map[string]int, len(result.Matrix))
	for i := range result.Matrix {
		data := util.ExtractValues(result.Matrix[i].Values)
		avgData, _ := util.CalculateAverage(data)
		label := result.Matrix[i].Metric.Cluster
		infoMap[label] = avgData
	}
	return infoMap
}
