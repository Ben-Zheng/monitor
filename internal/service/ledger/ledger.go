package ledger

import (
	"context"
	"github.com/jinzhu/copier"
	"log"
	"monitor/internal/service/excel"
	"monitor/internal/service/gpu"
)

//大模型调用情况台账
//智能平台服务调用情况
//大模型及其场景应用明细

type LedgerData struct {
	Ctx         context.Context
	modelLedger *ModelLedger
	sceneLedger *SceneLedger
	from        int64
	to          int64
}

func NewLedgerData(ctx context.Context) *LedgerData {
	modelLedger := NewModelLedger()
	sceneLedger := NewSceneLedger(ctx)
	return &LedgerData{
		Ctx:         ctx,
		modelLedger: modelLedger,
		sceneLedger: sceneLedger,
	}
}

type LargeModelServiceResp struct {
	ModelName        string `json:"modelName"`
	ApplyConcurrency int64  `json:"applyConcurrency"`
	InvokingCount    int64  `json:"InvokingCount"`
}

type InteResp struct {
	CallModelName      string `json:"callModelName"`
	AppScenarioName    string `json:"appScenarioName"`
	ApisixScenarioName string `json:"apisixScenarioName"`
	CallModelId        string `json:"callModelId"`
	DevDept            string `json:"devDept"`
	DevManager         string `json:"devManager"`
	EnvAlias           string `json:"envAlias"`
	EnvName            string `json:"envName"`
	ModelName          string `json:"modelName"`
	Token              string `json:"token"`
	MaxConcurrency     int64  `json:"maxConcurrency"`
	CountInvoking      int64  `json:"countInvoking"`
	TotalInvoking      int64  `json:"totalInvoking"`
	InvokingLastWeek   int64  `json:"invokingLastWeek"`
	InvokingThisWeek   int64  `json:"invokingThisWeek"`
	InvokingSuccess    int64  `json:"invokingSuccess"`
	InvokingChange     int64  `json:"invokingChange"`
	InvokingThisPeriod int64  `json:"invokingThisPeriod"`
	InvokingHistory    int64  `json:"invokingHistory"`
	ApplyModel         string `json:"applyModel"`
}

type ModelLedgerResp struct {
	TotalPvalue    float64
	ModelName      string //显卡型号
	NodeNum        int
	Corenum        int
	Pvalue         float64
	Model          string //模型名称
	UsedPvalue     float64
	UsedCards      int
	MaxConcurrency int64
}

// 算力分布情况台账
func (l *LedgerData) GetModelComputingDetail(from, to int64) ([]ModelLedgerResp, error) {
	queryStr := gpu.NewQueryExpr()
	details, err := queryStr.Getmodel_node_view(l.Ctx, from, to)

	if err != nil {
		return nil, err
	}

	GPU_Pvalue, err := queryStr.GetTotalNvidiaPvalue(l.Ctx, from, to)
	NPU_Pvalue, err := queryStr.GetTotalAscendPvalue(l.Ctx, from, to)
	sceneMap, err := l.sceneLedger.GetSceneInfoMap("model")

	modelsLedgerResp := make([]ModelLedgerResp, 0)
	for i := range details {
		var modelLedgerResp ModelLedgerResp
		detail := details[i]
		modelName := detail.Label_llm_model
		usedCards := detail.Core
		usedPvalue := detail.Pvalue
		label := detail.Resource

		if v, exist := GPU_Pvalue[label]; exist {
			modelLedgerResp.Corenum = v.Cards
			modelLedgerResp.NodeNum = v.NodesNum
			modelLedgerResp.Pvalue = v.Pvalue
		} else if v, exist := NPU_Pvalue[label]; exist {
			modelLedgerResp.Corenum = v.Cards
			modelLedgerResp.NodeNum = v.NodesNum
			modelLedgerResp.Pvalue = v.Pvalue
		}
		maxConcurr := sceneMap[modelName].MaxConcurrency
		modelLedgerResp.MaxConcurrency = maxConcurr
		modelLedgerResp.Model = modelName
		modelLedgerResp.UsedCards = usedCards
		modelLedgerResp.UsedPvalue = usedPvalue
		modelLedgerResp.ModelName = label
		modelsLedgerResp = append(modelsLedgerResp, modelLedgerResp)
	}
	return modelsLedgerResp, nil
}

// 大模型调用量情况
func (l *LedgerData) MakeLedgerLargeModel(from int64, to int64) ([]LargeModelServiceResp, error) {
	//获取数据
	sceneManager, err := l.sceneLedger.GetSceneInfoMap("model")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	modelInfo, err := l.modelLedger.GetSuccessInvokingByModelScene(from, to)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	modelCount := make([]LargeModelServiceResp, 0)
	for model := range modelInfo {
		var countScene int64
		var lmResp LargeModelServiceResp
		sceneMap := modelInfo[model]
		for scene := range sceneMap {
			countScene = countScene + sceneMap[scene]
		}
		lmResp.ModelName = model
		lmResp.ApplyConcurrency = sceneManager[model].MaxConcurrency
		lmResp.InvokingCount = countScene
		modelCount = append(modelCount, lmResp)
	}

	return modelCount, nil
}

func (l *LedgerData) MakeLedgerIntelligent(from int64, to int64) ([]InteResp, error) {
	//获取数据
	sceneManager, err := l.sceneLedger.GetSceneInfoMap("model")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	InvokingMap, err := l.modelLedger.GetSuccessInvokingByModelScene(from, to)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	IntelResps := make([]InteResp, 0)
	for k, v := range sceneManager {
		var intellResp InteResp
		token := v.Token
		//当前为mock

		countInvoking := InvokingMap[k]
		copier.Copy(&intellResp, v)
		intellResp.CountInvoking = countInvoking[token]
		IntelResps = append(IntelResps, intellResp)
	}
	return IntelResps, nil
}

func (l *LedgerData) MakeLedgerLargeModelDetail(from int64, to int64) ([]InteResp, error) {
	//获取数据
	sceneManager, err := l.sceneLedger.GetSceneInfoMap("scene")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println(sceneManager)
	//上周调用
	LastWeek_sceneInfo, err := l.modelLedger.GetLastWeekInvokingBySceneModel()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//本期调用
	sceneInfo, err := l.modelLedger.GetSuccessInvokingBySceneModel(from, to)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//var (
	//	authValues  []string
	//	modelValues []string
	//)
	//
	//for scene, modelInfo := range sceneInfo {
	//	for m := range modelInfo {
	//		authValues = append(authValues, scene)
	//		modelValues = append(modelValues, m)
	//		r := fmt.Sprintf("%s|%s", scene, m)
	//		fmt.Println(r)
	//	}
	//}
	//
	//history_sceneInfo, err := l.modelLedger.GetHistoryInvokingBySceneModel(authValues, modelValues)
	lmResps := make([]InteResp, 0)
	for scene, _ := range sceneManager {
		var lmResp InteResp
		modelInfo := sceneInfo[scene]
		for model := range modelInfo {
			lmResp.ModelName = model
			lmResp.Token = scene
			lmResp.ApisixScenarioName = sceneManager[scene].ApisixScenarioName
			lmResp.DevManager = sceneManager[scene].DevManager
			lmResp.EnvAlias = sceneManager[scene].EnvAlias
			lmResp.EnvName = sceneManager[scene].EnvName
			lmResp.ModelName = sceneManager[scene].ModelName
			lmResp.MaxConcurrency = sceneManager[scene].MaxConcurrency
			lmResp.InvokingLastWeek = LastWeek_sceneInfo[scene][model]
			lmResp.InvokingThisPeriod = modelInfo[model]
			lmResp.InvokingChange = lmResp.InvokingThisPeriod - lmResp.InvokingLastWeek

			//key := fmt.Sprintf("%s|%s", scene, model)
			//historycount := history_sceneInfo[key]
			//lmResp.InvokingHistory = historycount
			lmResps = append(lmResps, lmResp)
		}

	}
	return lmResps, nil
}

// 生成下列的四种台账
// 1、高性能算力及大模型部署情况
func (l *LedgerData) MakeHighLevelModelDetail(from, to int64) ([]excel.DataRow, error) {
	ComputingModel, err := l.GetModelComputingDetail(from, to)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	datas := make([]excel.DataRow, 0)
	for i := range ComputingModel {
		detail := ComputingModel[i]

		var data excel.DataRow
		data.Model = detail.ModelName
		data.ModelUsed = detail.Model
		data.CardNum = detail.Corenum
		data.ComputeP = detail.TotalPvalue
		data.ServerNum = detail.NodeNum
		data.UsedCompute = detail.UsedPvalue
		data.UsedCard = detail.UsedCards

		datas = append(datas, data)
	}

	return datas, nil
}
func (l *LedgerData) MakeLargeInvokingDetail(from, to int64) ([]excel.ServiceRecord, error) {
	LedgerInfo, err := l.MakeLedgerIntelligent(from, to)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	datas := make([]excel.ServiceRecord, 0)
	for i := range LedgerInfo {
		detail := LedgerInfo[i]
		var data excel.ServiceRecord
		data.Model = detail.ModelName
		data.Scene = detail.ApisixScenarioName
		data.Department = detail.DevDept
		data.ResponsiblePerson = detail.DevManager
		data.CallVolume = int(detail.InvokingThisPeriod)
		data.Concurrency = int(detail.MaxConcurrency)
		data.ApplyModel = detail.ApplyModel
		data.SerialNumber = i
		data.Environment = detail.EnvName
		data.Frequency = "实时"
		datas = append(datas, data)
	}
	return datas, nil
}
func (l *LedgerData) MakeplatformDetail(from, to int64) ([]excel.Record, error) {
	LedgerInfo, err := l.MakeLedgerLargeModelDetail(from, to)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	datas := make([]excel.Record, 0)
	for i := range LedgerInfo {
		detail := LedgerInfo[i]
		var data excel.Record
		data.Model = detail.ModelName
		data.Environment = detail.EnvName
		data.Department = detail.DevDept
		data.Success = int(detail.InvokingThisPeriod)
		data.Manager = detail.DevManager
		data.Scenario = detail.ApisixScenarioName
		datas = append(datas, data)
	}

	return datas, nil
}
