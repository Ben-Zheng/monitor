package model

import (
	"context"
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

// 这里负责获取日志部分代码
type ModelRepo interface {
	ModelsCountCards(req models.ModelsListRequest) (*types.PagedModelsResponse, error)
	ModelCountWithLog(req models.ModelWithCodeRequest) (*types.LogDetailResp, error)
	ModelRequestTime(req models.ModelWithCodeRequest) (*types.ModelRequestTimeResp, error)
	ModelsDetailInfo(req models.ModelWithCodeRequest) (*types.ModelDetailResp, error)
	ModelsDetailTrend(req models.ModelWithCodeRequest) (*types.ModelDetailTrend, error)
}

type ModelDomain struct {
	SceneMap map[string]string `json:"scene_map"`
	ctx      context.Context
}

func NewModelRepo(ctx context.Context, SceneMap map[string]string) ModelRepo {
	return &ModelDomain{
		SceneMap: SceneMap,
		ctx:      ctx,
	}
}

func (s *ModelDomain) ModelsDetailInfo(req models.ModelWithCodeRequest) (*types.ModelDetailResp, error) {
	appConfig := config.GetEsConfig()
	ec := es.NewESService(appConfig)
	from := util.ToInt64(req.From)
	to := util.ToInt64(req.To)
	timeDifferenceHours := float64(to-from) / (1000 * 60 * 60)
	resultScene, err := ec.CountSceneWithModel(from, to, req.ModelName)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	resultScenes := make([]types.SceneCountData, 0)
	for k := range resultScene {
		SceneName, exist := s.SceneMap[k]
		if !exist {
			continue
		}

		var sceneData types.SceneCountData
		useTimes := resultScene[k]
		diff_scene := float64(useTimes) / timeDifferenceHours
		sceneData.Qps = math.Round(diff_scene*100) / 100
		sceneData.SceneName = SceneName
		sceneData.SceneLabel = k
		sceneData.SceneCount = int(useTimes)
		resultScenes = append(resultScenes, sceneData)
	}

	return &types.ModelDetailResp{SceneDetails: resultScenes}, nil
}

func (s *ModelDomain) ModelsDetailTrend(req models.ModelWithCodeRequest) (*types.ModelDetailTrend, error) {
	appConfig := config.GetEsConfig()
	ec := es.NewESService(appConfig)
	resultTrend, err := ec.CountDailyLogsByFixedAuthModel(req.ModelName, req.AuthCode)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &types.ModelDetailTrend{ModelDetailTrend: resultTrend}, nil
}

func (s *ModelDomain) ModelsCountCards(req models.ModelsListRequest) (*types.PagedModelsResponse, error) {
	appConfig := config.GetEsConfig()
	ec := es.NewESService(appConfig)
	from := util.ToInt64(req.From)
	to := util.ToInt64(req.To)
	timeDifferenceHours := float64(to-from) / (1000 * 60 * 60)
	resultScence, err := ec.CountByModel(from, to)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	resultSceneData := make([]models.ModelCard, 0)
	gserver := NewModelsServer(s.ctx, from, to)
	for k, v := range resultScence {
		modelPodInfo := gserver.GetPods(k)
		var count int64
		sceneNum := len(v)
		for _, ts := range v {
			count += ts

		}
		var data models.ModelCard
		data.ModelName = k
		data.Scene = sceneNum
		data.Invoking = count
		diff := float64(count) / timeDifferenceHours
		data.Qps = math.Round(diff*100) / 100
		data.Status = modelPodInfo.status
		data.Url = modelPodInfo.url
		resultSceneData = append(resultSceneData, data)
	}
	totalItems := len(resultSceneData)

	sort.Slice(resultSceneData, func(i, j int) bool {
		return strings.ToLower(resultSceneData[i].ModelName) < strings.ToLower(resultSceneData[j].ModelName)
	})
	if req.Size == 0 {
		req.Size = 20
	}
	totalPages := totalItems / req.Size
	if totalItems%req.Size != 0 {
		totalPages++
	}

	hasNext := req.Page < totalPages
	return &types.PagedModelsResponse{
		Page:       req.Page,
		PageSize:   req.Size,
		TotalPages: totalPages,
		TotalItems: totalItems,
		HasNext:    hasNext,
		Data:       resultSceneData,
	}, nil
}

func (s *ModelDomain) ModelCountWithLog(req models.ModelWithCodeRequest) (*types.LogDetailResp, error) {
	appConfig := config.GetEsConfig()
	ec := es.NewESService(appConfig)
	from := util.ToInt64(req.From)
	to := util.ToInt64(req.To)

	resultScence, err := ec.GetDocumentFields(from, to, "failed", req.AuthCode, req.ModelName)
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
		authCode := util.ToString(resultScence[i]["http_authorization"], "")
		scene := s.SceneMap[authCode]
		Log := []string{_id, req_uri, time, path, request_time, request, http_model, http_host, url, scene}
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

// 统计响应时间，已废弃
func (s *ModelDomain) ModelRequestTime(req models.ModelWithCodeRequest) (*types.ModelRequestTimeResp, error) {
	appConfig := config.GetEsConfig()
	ec := es.NewESService(appConfig)
	from := util.ToInt64(req.From)
	to := util.ToInt64(req.To)
	resultScence, err := ec.GetDocumentFields(from, to, "success", "", req.ModelName)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	data := &types.ModelRequestTimeResp{}
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
		fmt.Println(err)
		return nil, err
	}
	requestTimeP50, requestTimeP90, requestTimeP95, err := util.CalculateP50P90P95(data.RequestTime)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	data.RequestTimeP50 = requestTimeP50
	data.RequestTimeP90 = requestTimeP90
	data.RequestTimeP95 = requestTimeP95

	data.UpstreamTimeP50 = upStreamP50
	data.UpstreamTimeP90 = upStreamP90
	data.UpstreamTimeP95 = upStreamP95

	data.ModelName = req.ModelName
	return data, nil
}
