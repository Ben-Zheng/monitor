package gpu

import (
	"fmt"
	"log"
	"monitor/internal/types"
	"monitor/util"
	"strings"
)

type HandlerPvalue struct {
	query   *QueryGrafana
	modeStr string
}

func NewHandlerPvalue(query *QueryGrafana) *HandlerPvalue {
	return &HandlerPvalue{
		query:   query,
		modeStr: query.ModeStr,
	}
}

func calPvalue(rule map[string]float64, core map[string]int) map[string]int {
	actualProduction := make(map[string]int)
	for t, production := range core {
		m := strings.ToLower(t)
		if multiplier, exists := rule[m]; exists {
			result := multiplier * float64(production)
			actualProduction[t] = int(result * 100)
		}
	}
	return actualProduction
}

func (q *HandlerPvalue) getClusterTotalPValue(clusterId string) int {

	nvidiaCore := q.getCountTotalByModelNameCluster("Nvidia", clusterId)
	ascendCore := q.getCountTotalByModelNameCluster("Ascend", clusterId)
	nvidiaPV := calPvalue(q.query.Rule, nvidiaCore)
	ascendPV := calPvalue(q.query.Rule, ascendCore)

	var pv int
	for _, v := range ascendPV {
		pv += v
	}

	for _, v := range nvidiaPV {
		pv += v
	}

	return pv
}

func (q *HandlerPvalue) getCountUsedcore(modelName, clusterId string) map[string]int {
	queryStr := NewQueryExpr()
	expr := queryStr.makeExprKpandaCluster(modelName, clusterId)
	result, err := q.query.Getinfo(expr)
	if err != nil {
		fmt.Println(err)
	}

	if result == nil || len(result.Matrix) == 0 {
		return make(map[string]int)
	}

	infoMap := make(map[string]int, len(result.Matrix))
	for i := range result.Matrix {
		data := util.ExtractValues(result.Matrix[i].Values)
		avgData, _ := util.CalculateMedian(data)
		label := result.Matrix[i].Metric.Cluster
		infoMap[label] += avgData
	}
	return infoMap
}

func (q *HandlerPvalue) getCountTotalByModelNameAllNode(modelName, clusterId string) map[string]int {
	queryStr := NewQueryExpr()
	expr := queryStr.makeTotalExpr(modelName, "node,modelName", "allnode", "count", clusterId, "")
	result, err := q.query.Getinfo(expr)
	if err != nil {
		fmt.Println(err)
	}
	totalCount := q.getMapInfoByModelNode(modelName, result)

	return totalCount
}

func (q *HandlerPvalue) getCountUsedByModelNameAllNode(modelName, clusterId string) map[string]int {
	queryStr := NewQueryExpr()
	expr := queryStr.makeTotalExpr(modelName, "node,modelName", "allnode", "count", clusterId, "")
	result, err := q.query.Getinfo(expr)
	if err != nil {
		fmt.Println(err)
	}

	exprUsed := queryStr.makeExprKpandaCluster("cluster", clusterId)
	resultUsedCore, err := q.query.Getinfo(exprUsed)
	if err != nil {
		fmt.Println(err)
	}

	infoMap := make(map[string]int, len(resultUsedCore.Matrix))
	for i := range resultUsedCore.Matrix {
		data := util.ExtractValues(resultUsedCore.Matrix[i].Values)
		avgData, _ := util.CalculateMedian(data)
		label := resultUsedCore.Matrix[i].Metric.Node
		infoMap[label] = avgData
	}

	usedCount := q.getMapInfoByModelNodeUsed(modelName, result, infoMap)

	return usedCount
}

func (q *HandlerPvalue) getCountTotalByModelNameCluster(modelName, clusterId string) map[string]int {
	queryStr := NewQueryExpr()
	expr := queryStr.makeTotalExpr(modelName, "modelName", "cluster", "count", clusterId, "")
	result, err := q.query.Getinfo(expr)
	if err != nil {
		fmt.Println(err)
	}
	totalCount := getMapInfoByModel(modelName, result)
	return totalCount
}

func (q *HandlerPvalue) getMapInfoByModelNode(modelName string, result *types.VectorResponse) map[string]int {

	if result == nil || len(result.Matrix) == 0 {
		return make(map[string]int)
	}

	mems := make(map[string]int, len(result.Matrix))
	for i := range result.Matrix {
		data := util.ExtractValues(result.Matrix[i].Values)
		mem, _ := util.CalculateAverage(data)
		var model string
		if modelName == "Ascend" {
			model = result.Matrix[i].Metric.Model_Name
		} else if modelName == "Nvidia" {
			model = result.Matrix[i].Metric.ModelName
		}

		model = strings.ToLower(model)
		node := result.Matrix[i].Metric.Node
		if q.query.Rule == nil {
			return make(map[string]int)
		}
		rule, exists := q.query.Rule[model]
		if !exists {
			log.Printf("模型 %s 的规则不存在", model)
			rule = 1.0 // 使用默认值
		}

		cal := int(rule * float64(mem) * 100)
		mems[node] = cal
	}

	return mems
}

func (q *HandlerPvalue) getMapInfoByModelNodeUsed(modelName string, result *types.VectorResponse, usedCore map[string]int) map[string]int {
	if result == nil || len(result.Matrix) == 0 {
		return make(map[string]int)
	}

	mems := make(map[string]int, len(result.Matrix))
	for i := range result.Matrix {
		data := util.ExtractValues(result.Matrix[i].Values)
		mem, _ := util.CalculateAverage(data)
		var model string
		if modelName == "Ascend" {
			model = result.Matrix[i].Metric.Model_Name
		} else if modelName == "Nvidia" {
			model = result.Matrix[i].Metric.ModelName
		}

		model = strings.ToLower(model)
		node := result.Matrix[i].Metric.Node

		cal := int(q.query.Rule[model] * float64(mem) * float64(usedCore[node]) / float64(mem) * 100)

		mems[node] = cal

	}

	return mems
}
