package api

import (
	"github.com/gin-gonic/gin"
	"log"
	"monitor/internal/common"
	"monitor/internal/models"
	"monitor/internal/service/model"
	"monitor/internal/service/scene"
	"monitor/internal/types"
	"net/http"
	"strconv"
	"time"
)

type ModelReq struct {
	iGrafanaService scene.IGrafanaService
}

func NewModelReq(iGrafanaService scene.IGrafanaService) *ModelReq {
	return &ModelReq{
		iGrafanaService: iGrafanaService,
	}
}

func (m *ModelReq) ModelCards(ctx *gin.Context) {
	result := &common.Result{}
	var params models.ModelsListRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}
	if params.From == "" || params.To == "" {
		now := time.Now()
		currentTimestamp := now.UnixMilli()
		fifteenDaysAgo := now.AddDate(0, 0, -30)
		fifteenDaysAgoTimestamp := fifteenDaysAgo.UnixMilli()
		params.To = strconv.FormatInt(currentTimestamp, 10)
		params.From = strconv.FormatInt(fifteenDaysAgoTimestamp, 10)
	}

	scnenLabel, err := m.iGrafanaService.GenerateApiSixScenarioKeyMap()
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "Internal service error"))
	}
	sl := model.NewModelRepo(ctx, scnenLabel)
	mdoelCardsResp, err := sl.ModelsCountCards(params)
	if err != nil {
		log.Println(err)
	}
	//if jsonData, err := json.MarshalIndent(result.Success(gin.H{"data": mdoelCardsResp}), "", "  "); err == nil {
	//	fmt.Println("JSON格式响应:")
	//	fmt.Println(string(jsonData))
	//} else {
	//	fmt.Println("JSON序列化失败:", err)
	//}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"data": mdoelCardsResp,
	}))
}

func (m *ModelReq) ModelReqTime(ctx *gin.Context) {
	result := &common.Result{}
	var params models.ModelWithCodeRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}
	if params.From == "" || params.To == "" {
		now := time.Now()
		currentTimestamp := now.UnixMilli()
		fifteenDaysAgo := now.AddDate(0, 0, -30)
		fifteenDaysAgoTimestamp := fifteenDaysAgo.UnixMilli()
		params.To = strconv.FormatInt(currentTimestamp, 10)
		params.From = strconv.FormatInt(fifteenDaysAgoTimestamp, 10)
	}

	if params.ModelName == "" {
		params.ModelName = "qwen"
	}

	scnenLabel, err := m.iGrafanaService.GenerateApiSixScenarioKeyMap()
	if err != nil {
		log.Println(err)
		ctx.JSON(500, result.Fail(http.StatusInternalServerError, "Internal service error"))
	}

	sl := model.NewModelRepo(ctx, scnenLabel)
	reqTimeResp, err := sl.ModelRequestTime(params)
	if err != nil {
		log.Println(err)
		ctx.JSON(500, result.Fail(http.StatusInternalServerError, "Internal service error"))
	}
	//if jsonData, err := json.MarshalIndent(result.Success(gin.H{"data": reqTimeResp}), "", "  "); err == nil {
	//	fmt.Println("JSON格式响应:")
	//	fmt.Println(string(jsonData))
	//} else {
	//	fmt.Println("JSON序列化失败:", err)
	//}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"data": reqTimeResp,
	}))
}

func (m *ModelReq) ModelDetail(ctx *gin.Context) {
	result := &common.Result{}
	var params models.ModelWithCodeRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}
	if params.From == "" || params.To == "" {
		now := time.Now()
		currentTimestamp := now.UnixMilli()
		monthDaysAgo := now.AddDate(0, 0, -30)
		monthDaysAgoTimestamp := monthDaysAgo.UnixMilli()
		params.To = strconv.FormatInt(currentTimestamp, 10)
		params.From = strconv.FormatInt(monthDaysAgoTimestamp, 10)
	}

	if params.ModelName == "" {
		params.ModelName = "qwen"
	}
	scnenLabel, err := m.iGrafanaService.GenerateApiSixScenarioKeyMap()
	sl := model.NewModelRepo(ctx, scnenLabel)

	modelDetailResp, err := sl.ModelsDetailInfo(params)
	if err != nil {
		log.Println(err)
		ctx.JSON(500, result.Fail(http.StatusInternalServerError, "Internal service error"))
	}

	modelLogResp, err := sl.ModelCountWithLog(params)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "Internal service error"))
	}

	var data types.ModelDetail
	data.Logs = modelLogResp
	data.Details = modelDetailResp

	//if jsonData, err := json.MarshalIndent(result.Success(gin.H{"data": data}), "", "  "); err == nil {
	//	fmt.Println("JSON格式响应:")
	//	fmt.Println(string(jsonData))
	//} else {
	//	fmt.Println("JSON序列化失败:", err)
	//}

	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"data": data,
	}))
}
