package ledger

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"log"
	"monitor/config"
	"monitor/internal/models"
	"monitor/internal/service/dao"
)

type LedgerClass int

const (
	// 1、高性能算力及大模型部署情况
	HighLevelLedgerClass LedgerClass = 1
	//2、大模型调用情况
	LargeModelLedgerClass LedgerClass = 2
	//3、大模型支撑场景调用量
	LargeModelSupportLedgerClass LedgerClass = 3
	//4、场景调用模型量明细
	SceneDetailLedgerClass LedgerClass = 4
)

// 返回台账excel给用户下载
// 创建任务
// 获取任务列表（缓存获取元数据）
// 一周按照7天时间戳计算

type TaskDomain struct {
	LedgerData *LedgerData
	Ctx        context.Context
	taskDao    dao.ITaskDao
}

type LedgerResult struct {
	Class LedgerClass
	Data  []interface{} // 实际数据
	Err   error         // 错误信息
}

// 获取台账task列表Get
// 查询台账任务Get
// 获取数据预览post（创建task）
// 生成台账post

func NewTaskDomain(ctx context.Context) *TaskDomain {
	cfg := config.GetDBConfig()
	db, err := dao.Connect(cfg)
	taskDao := dao.NewTokenDao(db)
	if err != nil {
		panic(err)
	}
	ledger := NewLedgerData(ctx)
	return &TaskDomain{
		taskDao:    taskDao,
		LedgerData: ledger,
	}
}

func (t *TaskDomain) GenerateTaskItem(params models.TaskMetaRequest) error {
	var taskMeta dao.TaskMetaData
	copier.Copy(&taskMeta, &params)
	taskMeta.DataType = params.LedgerType
	err := t.taskDao.CreateTask(&taskMeta)
	if err != nil {
		return err
	}
	return nil
}

// 获取任务列表
func (t *TaskDomain) GetTaskItems(name string, page int, pageSize int) ([]dao.TaskMetaData, int64, error) {
	//从缓存中获取到task元数据列表
	tasks, total, err := t.taskDao.GetTasksByName(name, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, err
}

// 获取台账任务
func (t *TaskDomain) TaskItems(page int, pageSize int) ([]dao.TaskMetaData, int64, error) {
	tasks, total, err := t.taskDao.GetTaskList(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, err
}

// 生成预览数据,返回数据,预览数据仅返回用户使用。
func (t *TaskDomain) GenerateLedgerData(ledgerclass LedgerClass, from, to int64) LedgerResult {
	switch ledgerclass {
	case HighLevelLedgerClass:

		data, err := t.LedgerData.MakeHighLevelModelDetail(from, to)
		if err != nil {
			log.Println(err)
		}
		var ledgerdata []interface{}
		for i := range data {
			ledgerdata = append(ledgerdata, data[i])
		}

		return LedgerResult{
			Class: HighLevelLedgerClass,
			Data:  ledgerdata,
			Err:   nil,
		}

	case LargeModelLedgerClass:
		fmt.Println("Large Model Ledger Class")
		data, err := t.LedgerData.MakeLargeInvokingDetail(from, to)
		if err != nil {
			log.Println(err)
		}
		var ledgerdata []interface{}
		for i := range data {
			ledgerdata = append(ledgerdata, data[i])
		}

		return LedgerResult{
			Class: LargeModelLedgerClass,
			Data:  ledgerdata,
			Err:   nil,
		}

	case LargeModelSupportLedgerClass:
		fmt.Println("Large Model Support Ledger Class")
		data, err := t.LedgerData.MakeLargeInvokingDetail(from, to)
		if err != nil {
			fmt.Println(err)
		}
		var ledgerdata []interface{}
		for i := range data {
			ledgerdata = append(ledgerdata, data[i])
		}
		return LedgerResult{
			Class: LargeModelSupportLedgerClass,
			Data:  ledgerdata,
			Err:   nil,
		}

	case SceneDetailLedgerClass:
		fmt.Println("Scene Detail Ledger Class")
		data, err := t.LedgerData.MakeplatformDetail(from, to)
		if err != nil {
			fmt.Println(err)
		}
		var ledgerdata []interface{}
		for i := range data {
			ledgerdata = append(ledgerdata, data[i])
		}
		return LedgerResult{
			Class: SceneDetailLedgerClass,
			Data:  ledgerdata,
			Err:   nil,
		}
	}
	return LedgerResult{}
}

// 生成台账,加入延时任务。获取post用户请求
func (t *TaskDomain) GenerateLedger() {

}
