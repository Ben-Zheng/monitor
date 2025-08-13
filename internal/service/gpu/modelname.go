package gpu

import (
	"fmt"
	"monitor/internal/types"
	"monitor/util"
)

func getMapInfoByModel(modelName string, result *types.VectorResponse) map[string]int {

	if result == nil || len(result.Matrix) == 0 {
		return make(map[string]int)
	}

	infoMap := make(map[string]int, len(result.Matrix))
	for i := range result.Matrix {
		data := util.ExtractValues(result.Matrix[i].Values)
		avgData, _ := util.CalculateAverage(data)
		var label string
		if modelName == "Ascend" {
			label = result.Matrix[i].Metric.Model_Name
		} else if modelName == "Nvidia" {
			label = result.Matrix[i].Metric.ModelName
		}
		infoMap[label] = avgData
	}
	return infoMap
}

type HandlerModelName struct {
	query *QueryGrafana
}

func NewHandlerModelName(query *QueryGrafana) *HandlerModelName {
	return &HandlerModelName{
		query: query,
	}
}

// 单个集群分型号的已用显存和所有显存
func (q *HandlerModelName) getUsedMemByModelNameCluster(modelName, clusterId string) (map[string]int, error) {
	queryStr := NewQueryExpr()
	expr := queryStr.makeUsedExpr(modelName, "modelName", "cluster", "sum", clusterId, "")
	result, err := q.query.Getinfo(expr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	usedMem := getMapInfoByModel(modelName, result)
	return usedMem, nil
}
func (q *HandlerModelName) getTotalMemByModelNameCluster(modelName, clusterId string) (map[string]int, error) {
	queryStr := NewQueryExpr()
	expr := queryStr.makeTotalExpr(modelName, "modelName", "cluster", "sum", clusterId, "")
	result, err := q.query.Getinfo(expr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	totalMem := getMapInfoByModel(modelName, result)
	return totalMem, nil
}
