package api

import (
	"github.com/gin-gonic/gin"
	"monitor/config"
	"monitor/internal/cache"
	"monitor/internal/client"
	"monitor/internal/common"
	"monitor/internal/models"
	"monitor/internal/service/gpu"
	"monitor/internal/types"
	"monitor/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ServiceContext struct {
	Cache *cache.SharedCache
}

func NewServiceContext() *ServiceContext {
	return &ServiceContext{
		Cache: cache.NewSharedCache(),
	}
}

func (s *ServiceContext) ListClusterName(ctx *gin.Context) {
	result := &common.Result{}
	clusters := gpu.GetClustersNames(ctx)
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"clusterNames": clusters,
	}))
}

func (s *ServiceContext) ListCluster(ctx *gin.Context) {

	result := &common.Result{}
	var params models.ClusterListRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	if params.From == "" {
		params.From = "now-5m"
	}
	if params.To == "" {
		params.To = "now"
	}
	if params.Page == 0 {
		params.Page = 1
	}

	if params.Size == 0 {
		params.Size = 10
	}
	baseTime := time.Now()
	from_timestamp, _ := util.ParseTimeInput(params.From, baseTime)
	to_timestamp, _ := util.ParseTimeInput(params.To, baseTime)
	info := gpu.GetClustersInfo(ctx, params, from_timestamp, to_timestamp)

	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"data": info,
	}))
}

func (s *ServiceContext) ListNodes(ctx *gin.Context) {
	result := &common.Result{}
	var params models.NodesListRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	if params.From == "" {
		params.From = "now-5m"
	}
	if params.To == "" {
		params.To = "now"
	}
	if params.Page == 0 {
		params.Page = 1
	}

	if params.Size == 0 {
		params.Size = 10
	}
	baseTime := time.Now()
	from_timestamp, _ := util.ParseTimeInput(params.From, baseTime)
	to_timestamp, _ := util.ParseTimeInput(params.To, baseTime)
	baseUrl := config.GetGrafanaQueryConfig().ClusterBaseURL
	info := gpu.GetNodesInfo(ctx, params, from_timestamp, to_timestamp, baseUrl)
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"data": info,
	}))
}

func (s *ServiceContext) ClusterDetail(ctx *gin.Context) {
	var result common.Result
	var params models.DetailRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}
	if params.From == "" {
		params.From = "now-5m"
	}
	if params.To == "" {
		params.To = "now"
	}

	baseTime := time.Now()
	from_timestamp, _ := util.ParseTimeInput(params.From, baseTime)
	to_timestamp, _ := util.ParseTimeInput(params.To, baseTime)

	fromStr := strconv.FormatInt(from_timestamp, 10)
	toStr := strconv.FormatInt(to_timestamp, 10)

	c := config.GetGrafanaQueryConfig()
	client := client.NewDCEClient(ctx, c.ClusterBaseURL, c.InsecureSkipVerify)
	cluNames := client.GetClusterName("/apis/insight.io/v1alpha1/metric/queryrange")
	var clusterName string
	for i := range cluNames {
		if cluNames[i].Cluster == params.Cluster {
			clusterName = cluNames[i].ClusterName
			break
		}
	}

	queries := map[string][]string{
		"from": {fromStr},
		"to":   {toStr},
	}

	vars := map[string][]string{
		"cluster_name": {clusterName},
		"node":         {"All"},
	}
	modeStr := gpu.GetClusterType(ctx, params.Cluster, from_timestamp, to_timestamp)

	var uid string
	var path string
	baseUrl := config.GetGrafanaQueryConfig().ClusterBaseURL
	if config.GetGrafanaQueryConfig().Mock == 1 {
		modeStr = "GPU"
	}
	if strings.ToLower(modeStr) == "npu" {
		uid = types.NpuUid
		path = types.NpuPath
		vars["npu"] = []string{"All"}

	} else if strings.ToLower(modeStr) == "gpu" {
		uid = types.GpuUid
		path = types.GpuPath
		vars["gpu"] = []string{"All"}

	}
	lang := common.LangEn

	// 调用函数构建 URL
	clusterInfo := gpu.GetClusterDetailByModel(ctx, params, from_timestamp, to_timestamp, baseUrl, modeStr)
	resPvalue := gpu.GetClusterDetailPValueByModel(ctx, params, from_timestamp, to_timestamp, baseUrl, modeStr, cluNames)
	coreInfo := gpu.GetCoresByCluster(ctx, params, from_timestamp, to_timestamp, baseUrl, cluNames)

	infoUrl := common.BuildGrafanaDashboardURL(uid, path, vars, queries, lang)
	modelVar := map[string][]string{
		"cluster_name": {clusterName},
		"mode":         {strings.ToUpper(modeStr)},
	}

	modelUrl := common.BuildGrafanaDashboardURL(types.ModelUid, types.ModelGrafanaPath, modelVar, queries, lang)
	var m models.DetailResponse
	m.Models = clusterInfo
	m.DetailUrl = infoUrl
	m.ModellUrl = modelUrl
	m.Pvalues = resPvalue
	m.UsedCore = coreInfo.UsedCore
	m.TotalCore = coreInfo.TotalCore
	m.TypeModel = strings.ToLower(modeStr)
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"data": m,
	}))
}

func (s *ServiceContext) NodeDetail(ctx *gin.Context) {
	var params models.DetailRequest
	var result common.Result
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	if params.From == "" {
		params.From = "now-5m"
	}

	if params.To == "" {
		params.To = "now"
	}

	var clustername string
	clusters := gpu.GetClustersNames(ctx)
	for i := range clusters {
		if clusters[i].Label == params.Cluster {
			clustername = clusters[i].Name
			break
		}
	}

	vars := map[string][]string{
		"node":         {params.Node},
		"cluster_name": {clustername},
	}

	baseTime := time.Now()
	from_timestamp, _ := util.ParseTimeInput(params.From, baseTime)
	to_timestamp, _ := util.ParseTimeInput(params.To, baseTime)

	modeStr := gpu.GetClusterType(ctx, params.Cluster, from_timestamp, to_timestamp)
	var uid string
	var path string
	var baseUrl string
	baseUrl = config.GetGrafanaQueryConfig().ClusterBaseURL
	if config.GetGrafanaQueryConfig().Mock == 1 {
		modeStr = "GPU"
	}
	if strings.ToLower(modeStr) == "npu" {
		uid = types.NodeNpuUid
		path = types.NodeNpuPath
		vars["npu"] = []string{"All"}
	} else if strings.ToLower(modeStr) == "gpu" {
		uid = types.NodeGpuUid
		path = types.NodeGpuPath
		vars["gpu"] = []string{"All"}
	}

	fromStr := strconv.FormatInt(from_timestamp, 10)
	toStr := strconv.FormatInt(to_timestamp, 10)
	queries := map[string][]string{
		"from": {fromStr},
		"to":   {toStr},
	}

	lang := common.LangEn
	url := common.BuildGrafanaDashboardURL(uid, path, vars, queries, lang)

	nodeInfo := gpu.GetNodesDetailByModel(ctx, params, from_timestamp, to_timestamp, baseUrl)
	Pvalues := gpu.GetNodesPvalueDetailByModel(ctx, params, from_timestamp, to_timestamp, baseUrl)
	coreInfo := gpu.GetCoresByNode(ctx, params, from_timestamp, to_timestamp, baseUrl)

	modelUid := types.NodeModelUid
	modelPath := types.NodeModelGrafanaPath

	modelVar := map[string][]string{
		"node":         {params.Node},
		"cluster_name": {clustername},
		"mode":         {modeStr},
	}

	modelUrl := common.BuildGrafanaDashboardURL(modelUid, modelPath, modelVar, queries, lang)
	var m models.NodeDetailResp
	m.Node = nodeInfo
	m.DetailUrl = url
	m.Pvalues = Pvalues
	m.ModellUrl = modelUrl
	m.UsedCore = coreInfo.UsedCore
	m.TotalCore = coreInfo.TotalCore
	m.TypeModel = strings.ToLower(modeStr)
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"data": m,
	}))
}
