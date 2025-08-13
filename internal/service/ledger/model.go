package ledger

import (
	"monitor/config"
	"monitor/internal/service/es"
	"monitor/util"
)

// 查询 model的本周调用次数
// 查询 场景+model的本周调用次数
// 查询 场景+model的累计调用次数
// 查询 场景+model的上周调用次数
// 查询 场景+model的本期调用次数

type ModelLedger struct {
	EsClient es.EsRepo
}

func NewModelLedger() *ModelLedger {
	appConfig := config.GetEsConfig()
	ec := es.NewESService(appConfig)
	return &ModelLedger{
		EsClient: ec,
	}
}

// [model]scene
func (m *ModelLedger) GetSuccessInvokingByModelScene(from int64, to int64) (map[string]map[string]int64, error) {
	return m.EsClient.Count(from, to, "success", "onModel", "")
}

// [scene]model
func (m *ModelLedger) GetSuccessInvokingBySceneModel(from int64, to int64) (map[string]map[string]int64, error) {
	return m.EsClient.Count(from, to, "success", "onAuth", "")
}

// [scene]model
func (m *ModelLedger) GetAllInvokingByModelScene(from int64, to int64) (map[string]map[string]int64, error) {
	return m.EsClient.Count(from, to, "", "onModel", "")
}

func (m *ModelLedger) GetLastWeekInvokingByModelScene() (map[string]map[string]int64, error) {
	from, to := util.GetLastWorkWeekTimestamps()
	return m.EsClient.Count(from, to, "", "onModel", "")
}

func (m *ModelLedger) GetLastWeekInvokingBySceneModel() (map[string]map[string]int64, error) {
	from, to := util.GetLastWorkWeekTimestamps()
	return m.EsClient.Count(from, to, "", "onAuth", "")
}

func (m *ModelLedger) GetHistoryInvokingBySceneModel(authValues []string, modelValues []string) (map[string]int64, error) {
	return m.EsClient.BatchCountFieldOccurrences(authValues, modelValues)
}
