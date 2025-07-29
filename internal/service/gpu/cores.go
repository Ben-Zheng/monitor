package gpu

import (
	"fmt"
	"monitor/util"
)

type HandlerCore struct {
	query   *QueryGrafana
	modeStr string
}

func NewHandlerCore(query *QueryGrafana) *HandlerCore {
	return &HandlerCore{
		query:   query,
		modeStr: query.ModeStr,
	}
}

func (q *HandlerCore) getAscendTotalCore() (map[string]int, error) {
	queryStr := NewQueryExpr()
	exprCores := queryStr.makeTotalExpr("Ascend", "cluster", "all", "count", "", "")
	result, err := q.query.Getinforange(exprCores)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	infoMap := make(map[string]int, len(result.Matrix))
	for i := range result.Matrix {
		data := util.ExtractValues(result.Matrix[i].Values)
		avgData, _ := util.CalculateMedian(data)
		label := result.Matrix[i].Metric.Cluster
		infoMap[label] = avgData
	}
	return infoMap, nil
}
func (q *HandlerCore) getNvidiaTotalCore() (map[string]int, error) {
	queryStr := NewQueryExpr()
	exprCores := queryStr.makeTotalExpr("Nvidia", "cluster", "all", "count", "", "")
	result, err := q.query.Getinfo(exprCores)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	infoMap := make(map[string]int, len(result.Matrix))
	for i := range result.Matrix {
		data := util.ExtractValues(result.Matrix[i].Values)
		avgData, _ := util.CalculateMedian(data)
		label := result.Matrix[i].Metric.Cluster
		infoMap[label] = avgData
	}
	return infoMap, nil
}

func (q *HandlerCore) getAscendNodesTotalMCore(clusterID string) (map[string]int, error) {
	number := fmt.Sprintf("count by (node) (npu_chip_info_hbm_total_memory{ cluster=\"%s\",node!=\"\"})", clusterID)
	result, err := q.query.Getinfo(number)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	infoMap := make(map[string]int, len(result.Matrix))
	for i := range result.Matrix {
		data := util.ExtractValues(result.Matrix[i].Values)
		avgData, _ := util.CalculateMedian(data)
		label := result.Matrix[i].Metric.Node
		infoMap[label] = avgData
	}
	return infoMap, nil

}
func (q *HandlerCore) getNvidiaNodesTotalMCore(clusterID string) (map[string]int, error) {
	number := fmt.Sprintf("count by (node) (DCGM_FI_DEV_FB_TOTAL{ cluster=\"%s\",node!=\"\"})", clusterID)
	result, err := q.query.Getinfo(number)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	infoMap := make(map[string]int, len(result.Matrix))
	for i := range result.Matrix {
		data := util.ExtractValues(result.Matrix[i].Values)
		avgData, _ := util.CalculateMedian(data)
		label := result.Matrix[i].Metric.Node
		infoMap[label] = avgData
	}

	return infoMap, nil
}

func (q *HandlerCore) getKpandaGpu() (map[string]int, error) {
	queryStr := NewQueryExpr()
	exprCores := queryStr.makeExprKpandaCluster("clusters", "")
	result, err := q.query.Getinfo(exprCores)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	infoMap := make(map[string]int, len(result.Matrix))
	for i := range result.Matrix {
		data := util.ExtractValues(result.Matrix[i].Values)
		avgData, _ := util.CalculateMedian(data)
		label := result.Matrix[i].Metric.Cluster
		infoMap[label] += avgData
	}
	return infoMap, nil
}
