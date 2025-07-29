package scene

import (
	"fmt"
	"log"
	"math"
	"monitor/config"
	"monitor/internal/common"
	"monitor/internal/models"
	"monitor/internal/service/es"
	"monitor/internal/types"
	"monitor/util"
	"sort"
	"strings"
)

type SceneRepo interface {
	ModelRequestTime(req models.SceneWithCodeRequest) (*types.RequestTimeResp, error)
	SceneCountCards(req models.SceneListRequest) (*types.ScenesPagedResponse, error)
	SceneCountWithModel(req models.SceneWithCodeRequest) (*types.SceneDetailWithModel, error)
	SceneCountWithLog(req models.SceneWithCodeRequest) (*types.LogDetailResp, error)
}

type SceneReq struct {
	SceneMap map[string]string `json:"scene_map"`
}

func NewSceneReq(SceneMap map[string]string) SceneRepo {
	return &SceneReq{
		SceneMap: SceneMap,
	}
}

// 响应时间
func (s *SceneReq) ModelRequestTime(req models.SceneWithCodeRequest) (*types.RequestTimeResp, error) {
	appConfig := config.GetEsConfig()
	ec := es.NewESService(appConfig)
	from := util.ToInt64(req.From)
	to := util.ToInt64(req.To)

	authCode := req.AuthCode
	resultScence, err := ec.GetDocumentFields(from, to, "success", authCode, req.ModelName)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	data := &types.RequestTimeResp{}
	for i := range resultScence {
		if resultScence[i]["request_time"] != nil {
			request_time := util.ToFloat64(resultScence[i]["request_time"], 0)
			data.RequestTime = append(data.RequestTime, request_time)
		}
		if resultScence[i]["upstream_response_time"] != nil {
			upstream_response_time := util.ToFloat64(resultScence[i]["upstream_response_time"], 0)
			data.UpstreamTime = append(data.UpstreamTime, upstream_response_time)
		}
	}

	upStreamP50, upStreamP90, upStreamP95, err := util.CalculateP50P90P95(data.UpstreamTime)
	if err != nil {
		log.Println(err)
	}

	requestTimeP50, requestTimeP90, requestTimeP95, err := util.CalculateP50P90P95(data.RequestTime)
	if err != nil {
		log.Println(err)
	}

	data.RequestTimeP50 = requestTimeP50
	data.RequestTimeP90 = requestTimeP90
	data.RequestTimeP95 = requestTimeP95

	data.UpstreamTimeP50 = upStreamP50
	data.UpstreamTimeP90 = upStreamP90
	data.UpstreamTimeP95 = upStreamP95

	data.SceneName, _ = s.SceneMap[authCode]
	data.SceneLabel = authCode

	return data, nil
}

// 卡片
func (s *SceneReq) SceneCountCards(req models.SceneListRequest) (*types.ScenesPagedResponse, error) {
	appConfig := config.GetEsConfig()
	ec := es.NewESService(appConfig)
	from := util.ToInt64(req.From)
	to := util.ToInt64(req.To)
	timeDifferenceHours := float64(to-from) / (1000 * 60 * 60)
	resultScence, err := ec.CountByNestedAggs(from, to)
	//resultScence, err := ec.GetDocumentFields(from, to, "all", "9s31a5kk574xsqwqnlw0wuxifn3ex6i7")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resultSceneData := make([]types.SceneCountData, 0)
	for k, v := range resultScence {
		var count int64
		SceneName := s.SceneMap[k]
		modelNum := len(v)
		for _, ts := range v {
			count += ts

		}
		var data types.SceneCountData
		data.SceneName = SceneName
		data.SceneCount = int(count)
		data.SceneLabel = k
		data.ModelsNum = modelNum
		diff := float64(count) / timeDifferenceHours
		data.Qps = math.Round(diff*100) / 100
		resultSceneData = append(resultSceneData, data)
	}
	totalItems := len(resultSceneData)

	sort.Slice(resultSceneData, func(i, j int) bool {
		return strings.ToLower(resultSceneData[i].SceneLabel) < strings.ToLower(resultSceneData[j].SceneLabel)
	})

	totalPages := totalItems / req.Size
	if totalItems%req.Size != 0 {
		totalPages++
	}

	hasNext := req.Page < totalPages
	return &types.ScenesPagedResponse{
		Page:       req.Page,
		PageSize:   req.Size,
		TotalPages: totalPages,
		TotalItems: totalItems,
		HasNext:    hasNext,
		Data:       resultSceneData,
	}, nil
}

func (s *SceneReq) SceneCountWithModel(req models.SceneWithCodeRequest) (*types.SceneDetailWithModel, error) {
	appConfig := config.GetEsConfig()
	ec := es.NewESService(appConfig)

	from := util.ToInt64(req.From)
	to := util.ToInt64(req.To)

	timeDifferenceHours := float64(to-from) / (1000 * 60 * 60)
	authCode := req.AuthCode
	resultScence, err := ec.CountByNestedAggs(from, to)
	Scene, _ := resultScence[authCode]
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	sceneResp := &types.SceneDetailWithModel{}
	sceneResp.SceneName = s.SceneMap[authCode]
	sceneResp.SceneLabel = authCode
	models := make([]types.ModelResp, 0)

	for k, v := range Scene {
		var data types.ModelResp
		data.Count = int(v)
		data.ModelName = k

		diff := float64(v) / timeDifferenceHours
		data.Qps = math.Round(diff*100) / 100

		models = append(models, data)
	}

	sceneResp.ModelsResp = models
	return sceneResp, nil
}

// 日志
func (s *SceneReq) SceneCountWithLog(req models.SceneWithCodeRequest) (*types.LogDetailResp, error) {
	appConfig := config.GetEsConfig()
	ec := es.NewESService(appConfig)
	from := util.ToInt64(req.From)
	to := util.ToInt64(req.To)

	resultScence, err := ec.GetDocumentFields(from, to, "failed", req.AuthCode, "")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	logDetails := make([]types.LogDetail, 0)
	for i := range resultScence {
		var data types.LogDetail
		id := util.ToString(resultScence[i]["_id"], "")
		kib := config.GetKibanaConfig()
		url, err := common.GenerateDocURL(kib, id)
		if err != nil {
			fmt.Println(err)
		}
		_id := fmt.Sprintf("_id: %s", util.ToString(resultScence[i]["_id"], ""))
		req_uri := fmt.Sprintf("request_uri: %s", util.ToString(resultScence[i]["request_uri"], ""))
		time := fmt.Sprintf("time: %s", util.ToString(resultScence[i]["time"], ""))
		path := fmt.Sprintf("path: %s", util.ToString(resultScence[i]["path"], ""))
		request_time := fmt.Sprintf("request_time: %s", util.ToString(resultScence[i]["request_time"], ""))
		http_model := fmt.Sprintf("http_model: %s", util.ToString(resultScence[i]["http_model"], ""))
		request := fmt.Sprintf("request: %s", util.ToString(resultScence[i]["request"], ""))
		http_host := fmt.Sprintf("http_host: %s", util.ToString(resultScence[i]["http_host"], ""))

		Log := []string{_id, req_uri, time, path, request_time, request, http_model, http_host, url}
		data.Log = Log

		if resultScence[i]["status"] != nil {
			data.Status = util.ToString(resultScence[i]["status"], "500")
		}

		logDetails = append(logDetails, data)
	}
	return &types.LogDetailResp{
		Logs: logDetails,
	}, nil
}
