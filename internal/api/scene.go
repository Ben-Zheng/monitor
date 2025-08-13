package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"monitor/internal/common"
	"monitor/internal/models"
	"monitor/internal/service/scene"
	"net/http"
	"strconv"
	"time"
)

type Scene struct {
	iGrafanaService scene.IGrafanaService
}

func NewScene(iGrafanaService scene.IGrafanaService) *Scene {
	return &Scene{
		iGrafanaService: iGrafanaService,
	}
}

func (s *Scene) CountScenes(ctx *gin.Context) {

	result := &common.Result{}
	var params models.SceneListRequest
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

	if params.Size == 0 {
		params.Size = 20
	}
	if params.Page == 0 {
		params.Page = 1
	}
	scnenLabel, err := s.iGrafanaService.GenerateApiSixScenarioKeyMap()
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "Internal service error"))
	}
	sl := scene.NewSceneReq(scnenLabel)
	cards, err := sl.SceneCountCards(params)

	ctx.JSON(http.StatusOK, result.Success(gin.H{"data": cards}))

}

func (s *Scene) CountModels(ctx *gin.Context) {

	result := &common.Result{}
	var params models.SceneWithCodeRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}
	if params.AuthCode == "" {
		ctx.JSON(http.StatusInternalServerError,
			result.Fail(http.StatusInternalServerError, "Internal service error"))
	}

	if params.From == "" || params.To == "" {
		now := time.Now()
		currentTimestamp := now.UnixMilli()
		fifteenDaysAgo := now.AddDate(0, 0, -30)
		fifteenDaysAgoTimestamp := fifteenDaysAgo.UnixMilli()
		params.To = strconv.FormatInt(currentTimestamp, 10)
		params.From = strconv.FormatInt(fifteenDaysAgoTimestamp, 10)
	}
	scnenLabel, err := s.iGrafanaService.GenerateApiSixScenarioKeyMap()
	sl := scene.NewSceneReq(scnenLabel)
	models, err := sl.SceneCountWithModel(params)
	detail, err := sl.SceneCountWithLog(params)
	detail.Models = models

	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "Internal service error"))
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"data": detail,
	}))
}

func (s *Scene) CountModelDetail(ctx *gin.Context) {
	result := &common.Result{}
	var params models.SceneWithCodeRequest
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

	scnenLabel, err := s.iGrafanaService.GenerateApiSixScenarioKeyMap()
	sl := scene.NewSceneReq(scnenLabel)

	reqTimeResp, err := sl.ModelRequestTime(params)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "Internal service error"))
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"data": reqTimeResp,
	}))
}
