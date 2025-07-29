package types

import (
	"monitor/internal/models"
	"time"
)

const (
	TimeLimit        = time.Hour
	ModelUid         = "gpu-cluster-useage"
	ModelGrafanaPath = "gpuji-qun-gpushi-yong-jian-kong"

	NodeModelUid         = "computing-node-details"
	NodeModelGrafanaPath = "suan-li-jie-dian-xiang-qing"

	NpuUid  = "gpu-cluster-overview"
	NpuPath = "gpuji-qun-gai-lan"
	GpuUid  = "gpu-cluster-overview"
	GpuPath = "gpuji-qun-gai-lan"

	NodeNpuUid  = "computing-node-overview"
	NodeNpuPath = "suan-li-jie-dian-gai-lan"
	NodeGpuUid  = "computing-node-overview"
	NodeGpuPath = "suan-li-jie-dian-gai-lan"
)

type SceneDetail struct {
	Id          string  `json:"_id"`
	RequestTime float64 `json:"request_time"`
	Status      int     `json:"status"`
	time        string  `json:"time"`
}

type SceneCountData struct {
	SceneCount int     `json:"scene_count"`
	ModelsNum  int     `json:"models_num"`
	SceneName  string  `json:"scene_name"`
	SceneLabel string  `json:"scene_label"`
	Qps        float64 `json:"qps"`
}

type RequestTimeResp struct {
	RequestTime  []float64 `json:"request_time"`
	UpstreamTime []float64 `json:"upstream_time"`

	RequestTimeP50 float64 `json:"request_time_p50"`
	RequestTimeP90 float64 `json:"request_time_p90"`
	RequestTimeP95 float64 `json:"request_time_p95"`

	UpstreamTimeP50 float64 `json:"upstream_time_p50"`
	UpstreamTimeP90 float64 `json:"upstream_time_p90"`
	UpstreamTimeP95 float64 `json:"upstream_time_p95"`

	SceneName  string `json:"scene_name"`
	SceneLabel string `json:"scene_label"`
}

type LogDetail struct {
	Status string   `json:"status"`
	Log    []string `json:"log"`
}
type LogDetailResp struct {
	Logs   []LogDetail           `json:"logs"`
	Models *SceneDetailWithModel `json:"models"`
}

type ModelResp struct {
	Count     int     `json:"count"`
	ModelName string  `json:"model_name"`
	Qps       float64 `json:"qps"`
}

type SceneDetailWithModel struct {
	ModelsResp []ModelResp `json:"models_resp"`
	SceneName  string      `json:"scene_name"`
	SceneLabel string      `json:"scene_label"`
}

// 定义最外层结构体
type VectorResponse struct {
	Matrix []MatrixItem `json:"matrix"`
}

// 定义 vector 数组中的每个元素
type MatrixItem struct {
	Metric Metric      `json:"metric"`
	Values []DataPoint `json:"values"`
}

// 定义 metric 对象
type Metric struct {
	Cluster     string `json:"cluster,omitempty"`
	ModelName   string `json:"modelname,omitempty"`
	Node        string `json:"node,omitempty"`
	Mode        string `json:"mode,omitempty"`
	ClusterName string `json:"clustername,omitempty"`
	// 别名字段
	Cluster_Name string `json:"cluster_name,omitempty"`
	Model_Name   string `json:"model_name,omitempty"`
	//模型名称
	Llm_model string `json:"llm_model,omitempty"`
	//podId
	Pod string `json:"pod,omitempty"`
}

// 数据点
type DataPoint struct {
	Timestamp string `json:"timestamp"`
	Value     string `json:"value"`
}

type ResponseNameList struct {
	Status    string     `json:"status"`
	IsPartial bool       `json:"isPartial"`
	Data      []NameList `json:"data"`
}

type NameList struct {
	Name        string `json:"__name__"`
	Cluster     string `json:"cluster"`
	ClusterName string `json:"cluster_name"`
	GPU         string `json:"gpu"`
}

type PagedResponse struct {
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
	TotalItems int              `json:"total_items"`
	HasNext    bool             `json:"has_next"`
	Data       []models.Cluster `json:"data"`
}

type ScenesPagedResponse struct {
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
	TotalItems int              `json:"total_items"`
	HasNext    bool             `json:"has_next"`
	Data       []SceneCountData `json:"data"`
}

type PagedNodesResponse struct {
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
	TotalItems int           `json:"total_items"`
	HasNext    bool          `json:"has_next"`
	Data       []models.Node `json:"data"`
}

// ------------------------------------------------
type PagedModelsResponse struct {
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
	TotalItems int                `json:"total_items"`
	HasNext    bool               `json:"has_next"`
	Data       []models.ModelCard `json:"data"`
}

type ModelRequestTimeResp struct {
	RequestTime  []float64 `json:"request_time"`
	UpstreamTime []float64 `json:"upstream_time"`

	RequestTimeP50 float64 `json:"request_time_p50"`
	RequestTimeP90 float64 `json:"request_time_p90"`
	RequestTimeP95 float64 `json:"request_time_p95"`

	UpstreamTimeP50 float64 `json:"upstream_time_p50"`
	UpstreamTimeP90 float64 `json:"upstream_time_p90"`
	UpstreamTimeP95 float64 `json:"upstream_time_p95"`
	ModelName       string  `json:"model_name"`
}

type ModelLogDetail struct {
	Status string   `json:"status"`
	Log    []string `json:"log"`
}

type ModelDetail struct {
	Logs    *LogDetailResp   `json:"logs"`
	Details *ModelDetailResp `json:"details"`
}
type ModelDetailResp struct {
	ModelDetail     models.ModelCard
	ModelPodDetails []models.ModelPodDetail
}
