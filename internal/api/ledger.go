package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"math"
	"monitor/internal/common"
	"monitor/internal/models"
	"monitor/internal/service/dao"
	"monitor/internal/service/excel"
	"monitor/internal/service/ledger"
	"monitor/internal/service/task"
	"monitor/internal/types"
	"monitor/util"
	"net/http"
	"os"
	"path/filepath"
)

type LedgerService struct {
	Domain        *ledger.TaskDomain
	TaskScheduler *task.DelayedTaskScheduler
}

func NewLedger(ctx context.Context) *LedgerService {
	domain := ledger.NewTaskDomain(ctx)
	taskScheduler := task.NewDelayedTaskScheduler()
	return &LedgerService{
		Domain:        domain,
		TaskScheduler: taskScheduler,
	}
}

// 任务列表
func (t *LedgerService) LedgerTasksList(ctx *gin.Context) {
	result := &common.Result{}
	var params models.TaskListRequest

	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}
	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Size <= 0 {
		params.Size = 10 // 默认每页10条
	}
	if params.Size > 100 {
		params.Size = 100 // 限制最大100条
	}
	tasks := make([]dao.TaskMetaData, 0)
	var total int64
	var err error
	if params.Name != "" {
		tasks, total, err = t.Domain.GetTaskItems(params.Name, params.Page, params.Size)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
			return
		}
	} else {
		tasks, total, err = t.Domain.TaskItems(params.Page, params.Size)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
			return
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Size)))

	// 计算是否有下一页
	hasNext := params.Page < totalPages
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}
	// 构建响应
	response := types.PagedResponseTask{
		Page:       params.Page,
		PageSize:   params.Size,
		HasNext:    hasNext,
		TotalPages: totalPages,
		TotalItems: int(total),
		Data:       tasks,
	}

	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"data": response,
	}))
}

// 数据预览
func (t *LedgerService) LedgerAllInfo(ctx *gin.Context) {
	result := &common.Result{}
	var params models.DataGenerateReq
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}
	ledgerType := ledger.LedgerClass(params.LedgerType)
	from := util.DayTomill(params.From)
	to := util.DayTomill(params.To)

	log.Println("LedgerAllInfo", ledgerType, from, to)

	info := t.Domain.GenerateLedgerData(ledgerType, from, to)
	log.Println("LedgerAllInfo", info)
	if info.Err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": info.Err.Error()})
	}

	switch ledgerType {
	case ledger.HighLevelLedgerClass:
		var records []excel.DataRow
		for _, item := range info.Data {
			if row, ok := item.(excel.DataRow); ok {
				records = append(records, row)
			}
		}
		ctx.JSON(http.StatusOK, result.Success(gin.H{
			"data": records,
		}))

	case ledger.LargeModelLedgerClass:
		var records []excel.ServiceRecord
		for _, item := range info.Data {
			if row, ok := item.(excel.ServiceRecord); ok {
				records = append(records, row)
			}
		}
		ctx.JSON(http.StatusOK, result.Success(gin.H{
			"data": records,
		}))

	case ledger.LargeModelSupportLedgerClass:
		var records []excel.ServiceRecord
		for _, item := range info.Data {
			if row, ok := item.(excel.ServiceRecord); ok {
				records = append(records, row)
			}
		}
		ctx.JSON(http.StatusOK, result.Success(gin.H{
			"data": records,
		}))

	case ledger.SceneDetailLedgerClass:
		var records []excel.Record
		for _, item := range info.Data {
			if row, ok := item.(excel.Record); ok {
				records = append(records, row)
			}
		}
		ctx.JSON(http.StatusOK, result.Success(gin.H{
			"data": records,
		}))

	}
	ctx.JSON(http.StatusBadRequest, gin.H{"error": info.Err.Error()})
}

// 台账生成
func (t *LedgerService) GenerateLedger(ctx *gin.Context) {
	result := &common.Result{}
	var params models.TaskMetaRequest

	if err := ctx.ShouldBind(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	log.Println("GenerateLedger", params)
	// 获得post的数据，并转成excel保存到本地
	ledgerType := ledger.LedgerClass(params.LedgerType)
	var filename string

	switch ledgerType {
	case ledger.HighLevelLedgerClass:
		var records []excel.DataRow
		if err := json.Unmarshal(params.Data, &records); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
			return
		}
		h := excel.NewHighLevel()
		filename = h.GenerateLedger(records)

	case ledger.LargeModelLedgerClass:
		var records []excel.ServiceRecord
		if err := json.Unmarshal(params.Data, &records); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
			return
		}
		h := excel.NewLargeInvokingexcel()
		filename = h.GenerateLedgerExcel(records)

	case ledger.LargeModelSupportLedgerClass:
		var records []excel.ServiceRecord
		if err := json.Unmarshal(params.Data, &records); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
			return
		}
		h := excel.NewLargeInvokingexcel()
		filename = h.GenerateLedgerExcel(records)

	case ledger.SceneDetailLedgerClass:
		var records []excel.Record
		if err := json.Unmarshal(params.Data, &records); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
			return
		}
		s := excel.NewServiceLedgerDetail()
		filename = s.GenerateServiceLedger(records)
	}
	var ledgerinfo types.GenerateLedgerResp
	ledgerinfo.LedgerName = filename
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"data": ledgerinfo,
	}))
}

// 任务生成
func (t *LedgerService) GenerateTask(ctx *gin.Context) {
	result := &common.Result{}
	var params models.TaskMetaRequest

	if err := ctx.ShouldBind(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	err := t.Domain.GenerateTaskItem(params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	// 是否需要发送邮件，如果是，加入延时任务
	//go t.TaskScheduler.Schedule("C", 10000*time.Millisecond, func(ctx context.Context) error {
	//	//TODO 发送邮件
	//	return nil
	//})
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"data": "success",
	}))
}
func (t *LedgerService) DownloadLedger(ctx *gin.Context) {

	var params models.DownloadLedgerReq
	if err := ctx.ShouldBind(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	filePath := fmt.Sprintf("./files/%s", params.LedgerName)
	// 2. 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "文件不存在",
			"path":  filePath,
		})
		return
	}

	// 3. 读取文件内容
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":  "读取文件失败",
			"detail": err.Error(),
		})
		return
	}

	fileName := filepath.Base(filePath)
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Cache-Control", "no-cache")

	ctx.Data(http.StatusOK, "application/octet-stream", fileBytes)
}
