package ledger

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"log"
	"monitor/config"
	"monitor/internal/client"
	"monitor/internal/service/scene"
	"monitor/internal/types"
	"strings"
)

const (
	Model         KeyModel = "model"
	Scene         KeyModel = "scene"
	CallModelName KeyModel = "callmodelname"
)

type SceneLedger struct {
	client *client.DCEClient
}
type KeyModel string

func NewSceneLedger(ctx context.Context) *SceneLedger {
	c := config.GetGrafanaQueryConfig()
	client := client.NewDCEClient(ctx, c.ClusterBaseURL, c.InsecureSkipVerify)
	return &SceneLedger{
		client: client,
	}
}

// 传参为"scene" key为token 传参为“model”key为model 传参为“callmodelname”key为模型描述。
func (s *SceneLedger) GetSceneInfoMap(key KeyModel) (map[string]types.SceneInfoItem, error) {
	url := "/apis/auth.engine.io/v1/workspaces/2/tokens/list"
	pageSize := 50
	infosResp, err := s.client.GetSceneManageInfo(url, pageSize)
	if err != nil {
		return nil, err
	}
	grafanaConf := config.Grafana
	iGrafanaService := scene.NewGrafanaService(&grafanaConf)
	sceneLabel, err := iGrafanaService.GenerateApiSixScenarioKeyMapMock()
	for k, v := range sceneLabel {
		fmt.Println("key", k)
		fmt.Println("value", v)
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}

	SceneMap := make(map[string]types.SceneInfoItem, 0)
	for i := range infosResp {
		var data types.SceneInfoItem
		err := copier.Copy(&data, infosResp[i])
		log.Println(data)
		if err != nil {
			return nil, err
		}
		if key == Scene {
			apisixScenarioName := strings.ToLower(data.ApisixScenarioName)
			token := sceneLabel[apisixScenarioName]
			SceneMap[token] = data
		} else if key == Model {
			SceneMap[data.ModelName] = data
		} else if key == CallModelName {
			SceneMap[data.CallModelName] = data
		}
	}

	return SceneMap, nil
}
