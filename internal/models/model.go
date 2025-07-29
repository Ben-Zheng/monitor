package models

type Cluster struct {
	Label          string  `json:"label"`
	Name           string  `json:"name"`
	GPUCores       int     `json:"gpuCores"`
	UsedGPUMemory  int     `json:"usedGPUMemory"`
	TotalGPUMemory int     `json:"totalGPUMemory"`
	UsedPValue     float64 `json:"usedPValue"`
	TotalPValue    float64 `json:"totalPValue"`
}

type Node struct {
	Name           string  `json:"name"`
	GPUCores       int     `json:"gpuCores"`
	UsedGPUMemory  int     `json:"usedGPUMemory"`
	TotalGPUMemory int     `json:"totalGPUMemory"`
	PvalueTotal    float64 `json:"usedPValue"`
	PvalueUsed     float64 `json:"totalPValue"`
}

type ClusterListRequest struct {
	From string `form:"from"`
	To   string `form:"to"`
	Page int    `form:"page"`
	Size int    `form:"size"`
}

type NodesListRequest struct {
	From    string `form:"from"`
	To      string `form:"to"`
	Page    int    `form:"page"`
	Size    int    `form:"size"`
	Cluster string `form:"cluster"`
}

type DetailRequest struct {
	ClusterName string `form:"cluster_name"`
	Cluster     string `form:"cluster"`
	Node        string `form:"node"`
	GPU         string `form:"gpu"`
	From        string `form:"from"`
	To          string `form:"to"`
}
type NodeDetailResp struct {
	DetailUrl string `json:"detail_url"`
	ModellUrl string `json:"model_url"`
	Node      NodeDetailResponse
	Pvalues   []ModelPvalueResponse
	UsedCore  int    `json:"used_core"`
	TotalCore int    `json:"total_core"`
	TypeModel string `json:"type_model"`
}

type DetailResponse struct {
	DetailUrl string `json:"detail_url"`
	ModellUrl string `json:"model_url"`
	Models    []ModelDetailResponse
	Pvalues   []ModelPvalueResponse
	UsedCore  int    `json:"used_core"`
	TotalCore int    `json:"total_core"`
	TypeModel string `json:"type_model"`
}

type CoreResponse struct {
	UsedCore  int `json:"used_core"`
	TotalCore int `json:"total_core"`
}
type ModelDetailResponse struct {
	NodeName    string `json:"node_name"`
	UsedMem     int    `json:"used_mem"`
	UsedPercent int    `json:"percent"`
	TotalMem    int    `json:"total_mem"`
}

type NodeDetailResponse struct {
	NodeName    string `json:"node_name"`
	TotalMem    int    `json:"total_mem"`
	UsedMem     int    `json:"used_mem"`
	UsedPercent int    `json:"percent"`
}
type ModelPvalueResponse struct {
	ModelName string  `json:"model_name"`
	Pvalue    float64 `json:"pvalue"`
}

type Item struct {
	Index int    `json:"index"`
	Text  string `json:"text"` // 场景
}

type FieldConfig struct {
	Defaults Defaults `json:"defaults"`
}

type Defaults struct {
	Mappings []Mapping `json:"mappings"`
}

type Mapping struct {
	Options map[string]Item `json:"options"`
}
type Panel struct {
	Title       string      `json:"title"`
	FieldConfig FieldConfig `json:"fieldConfig"`
}

type ModelCard struct {
	ModelName string  `json:"model_name"`
	Scene     int     `json:"scene"`    //场景绑定数量
	Invoking  int64   `json:"invoking"` //模型调用次数
	Qps       float64 `json:"qps"`      //模型qps
}

type ModelPodDetail struct {
	Pod         string  `json:"pod"`
	ClusterName string  `json:"cluster_name"`
	Cluster     string  `json:"cluster"`
	MemUsage    float64 `json:"mem_usage"`
	CpuUsage    float64 `json:"cpu_usage"`
	PodStatus   string  `json:"pod_status"`
	ModelName   string  `json:"model_name"`
}

type ClusterPod struct {
	ClusterName string `json:"cluster_name"`
	Cluster     string `json:"cluster"`
	PodInfoList []ModelPodDetail
}
type SceneListRequest struct {
	From string `form:"from"`
	To   string `form:"to"`
	Page int    `form:"page"`
	Size int    `form:"size"`
}

type SceneWithCodeRequest struct {
	From      string `form:"from"`
	To        string `form:"to"`
	AuthCode  string `form:"authorization_code"`
	ModelName string `form:"model_name"`
}

type ModelsListRequest struct {
	From string `form:"from"`
	To   string `form:"to"`
	Page int    `form:"page"`
	Size int    `form:"size"`
}
type ModelWithCodeRequest struct {
	From      string `form:"from"`
	To        string `form:"to"`
	AuthCode  string `form:"authorization_code"`
	ModelName string `form:"model_name"`
}
