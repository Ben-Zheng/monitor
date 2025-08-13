package gpu

import (
	"fmt"
	"log"
	"monitor/config"
	"monitor/internal/client"
	"monitor/internal/types"
	"monitor/util"
	"strconv"
)

type HandlerNode struct {
	query   *QueryGrafana
	modeStr string
}

func NewHandlerNode(query *QueryGrafana, clusterId string) *HandlerNode {
	modeStr := query.GetModeStr(clusterId)
	return &HandlerNode{
		query:   query,
		modeStr: modeStr,
	}
}

func (q *HandlerNode) getNodesName(clusterID string) []NodeInfo {
	qNvidiaStr := fmt.Sprintf("count by (node)(DCGM_FI_DEV_FB_TOTAL{cluster=\"%s\"})", clusterID)
	qAscendStr := fmt.Sprintf("count by (node)(npu_chip_info_hbm_total_memory{cluster=\"%s\"})", clusterID)
	Nnodes := q.getNames(qNvidiaStr)
	Anodes := q.getNames(qAscendStr)
	return mergeAndDeduplicate(Nnodes, Anodes)
}

func (q *HandlerNode) getNames(qStr string) []NodeInfo {

	query := config.NewQueryConfig()
	start := strconv.FormatInt(q.query.From, 10)
	end := strconv.FormatInt(q.query.To, 10)
	queryStr := query.BuildGrafanaQueryRange(qStr, start, end)
	client := client.NewDCEClient(q.query.Ctx, q.query.Url, true)
	result, err := client.MakeGetReqRange(queryStr, "/apis/insight.io/v1alpha1/metric/queryrange")

	if err != nil {
		fmt.Println(err)
	}

	clusters := make([]NodeInfo, len(result.Matrix))
	for i := range clusters {
		clusters[i].Name = result.Matrix[i].Metric.Node
	}
	return clusters
}

// 单个节点的所有显存和已用显存
func (q *HandlerNode) getTotalMemByNode(modelName string, clusterId string, nodeId string) (map[string]int, error) {
	queryStr := NewQueryExpr()
	expr := queryStr.makeTotalExpr(modelName, "cluster,node", "node", "sum", clusterId, nodeId)
	result, err := q.query.Getinfo(expr)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	usedMem := getMapInfoNode(result)
	return usedMem, nil
}

func (q *HandlerNode) getUsedMemByNode(modelName string, clusterId string, nodeId string) (map[string]int, error) {
	queryStr := NewQueryExpr()
	expr := queryStr.makeUsedExpr(modelName, "cluster,node", "node", "sum", clusterId, nodeId)
	result, err := q.query.Getinfo(expr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	usedMem := getMapInfoNode(result)
	return usedMem, nil
}

func (q *HandlerNode) getUsedMemByNodeAll(modelName string, clusterId string) (map[string]int, error) {
	queryStr := NewQueryExpr()
	expr := queryStr.makeUsedExpr(modelName, "node", "allnode", "sum", clusterId, "")
	result, err := q.query.Getinfo(expr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	usedMem := getMapInfoNode(result)
	return usedMem, nil
}

func (q *HandlerNode) getTotalMemByNodeAll(modelName string, clusterId string) (map[string]int, error) {
	queryStr := NewQueryExpr()
	expr := queryStr.makeTotalExpr(modelName, "node", "allnode", "sum", clusterId, "")
	result, err := q.query.Getinfo(expr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	usedMem := getMapInfoNode(result)
	return usedMem, nil
}

func (q *HandlerNode) getUsedPvalueByNode(modelName string, clusterId string, nodeId string) (map[string]int, error) {
	queryStr := NewQueryExpr()
	expr := queryStr.makeTotalExpr(modelName, "modelName", "node", "count", clusterId, nodeId)
	result, err := q.query.Getinfo(expr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var exprUsed string
	//[modelName]totalPvalue
	totalMem := getMapInfoByModel(modelName, result)
	if modelName == "Nvidia" {
		exprUsed = queryStr.makeExprKpandaNode(clusterId, nodeId)
	} else if modelName == "Ascend" {
		exprUsed = queryStr.makeExprKpandaNode(clusterId, nodeId)
	}

	resultUsedCore, err := q.query.Getinfo(exprUsed)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	infoMap := make(map[string]int)
	for i := range resultUsedCore.Matrix {
		data := util.ExtractValues(resultUsedCore.Matrix[i].Values)
		avgData, _ := util.CalculateMedian(data)
		label := resultUsedCore.Matrix[i].Metric.Node
		infoMap[label] = avgData
	}
	totalMemP := calPvalue(q.query.Rule, totalMem)
	//[modetal]pvalue
	usedMemP := make(map[string]int)
	for k, v := range totalMemP {
		percent := int(float64(infoMap[nodeId]) / float64(totalMem[k]) * 100)
		usedMemP[nodeId] = v * percent / 100
	}
	return usedMemP, nil
}

func (q *HandlerNode) getUsedCoreByNode(modelName string, clusterId string, nodeId string) (map[string]int, error) {
	queryStr := NewQueryExpr()
	expr := queryStr.makeUsedExpr(modelName, "node", "node", "count", clusterId, nodeId)
	result, err := q.query.Getinfo(expr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	usedMem := getMapInfoNode(result)
	return usedMem, nil
}

func (q *HandlerNode) getTotalCoreByNode(modelName string, clusterId string, nodeId string) (map[string]int, error) {
	queryStr := NewQueryExpr()
	expr := queryStr.makeTotalExpr(modelName, "node", "node", "count", clusterId, nodeId)

	result, err := q.query.Getinfo(expr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	usedMem := getMapInfoNode(result)
	return usedMem, nil
}

func getMapInfoNode(result *types.VectorResponse) map[string]int {

	if result == nil || len(result.Matrix) == 0 {
		return make(map[string]int)
	}

	infoMap := make(map[string]int, len(result.Matrix))
	for i := range result.Matrix {
		data := util.ExtractValues(result.Matrix[i].Values)
		avgData, _ := util.CalculateAverage(data)
		label := result.Matrix[i].Metric.Node
		infoMap[label] = avgData
	}
	return infoMap
}
func mergeAndDeduplicate(slice1, slice2 []NodeInfo) []NodeInfo {
	uniqueNames := make(map[string]bool)
	var result []NodeInfo

	for _, node := range append(slice1, slice2...) {
		if _, exists := uniqueNames[node.Name]; !exists {
			uniqueNames[node.Name] = true
			result = append(result, node)
		}
	}

	return result
}
