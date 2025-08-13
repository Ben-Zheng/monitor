package excel

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"log"
	"monitor/util"
	"strconv"
)

// ExcelColumn 定义Excel表格列
type ExcelColumn struct {
	Name  string
	Merge bool
}

// 表格列配置
var columns = []ExcelColumn{
	{"环境", true},
	{"算力(P)", true}, // 算力列也需要合并
	{"型号", true},
	{"服务器数(台)", false},
	{"卡数(张)", false},
	{"算力(P)", false},
	{"模型", false},
	{"使用算力", false},
	{"使用卡数", false},
	{"备注", false},
}

// DataRow 定义数据行结构
type DataRow struct {
	Env         string  `json:"env"`         //环境
	ComputeP    float64 `json:"computeP"`    // 算力列
	Model       string  `json:"model"`       // 显卡型号
	ServerNum   int     `json:"serverNum"`   //服务器数量
	CardNum     int     `json:"cardNum"`     //显卡张数
	ModelUsed   string  `json:"modelUsed"`   //模型名称
	UsedCompute float64 `json:"usedCompute"` //使用算力
	UsedCard    int     `json:"usedCard"`    //使用卡数
	Remarks     string  `json:"remarks"`     //备注
}
type HighLevel struct {
}

func NewHighLevel() *HighLevel {
	return &HighLevel{}
}
func (h *HighLevel) GenerateLedger(newData []DataRow) string {
	// 新数据示例（按环境分组，相同型号连续）
	//newData = []DataRow{
	//	{"生产", 461.5, "910B", 175, 1400, "DeepSeek-R1-Distill-Qwen-32B", 45, 144, ""},
	//	{"生产", 461.5, "910B", 175, 1400, "QwQ-32B", 7.5, 24, ""},
	//	{"生产", 461.5, "910B", 175, 1400, "Qwen2.5-v1-32B", 8.7, 27, ""},
	//	{"生产", 461.5, "910B", 175, 1400, "Qwen3-32B", 0.3, 1, ""},
	//	{"生产", 521.5, "v300", 175, 1400, "Qwen2-instruct-72B", 264, 848, ""},
	//	{"生产", 521.5, "v300", 175, 1400, "Qwen3-8B", 25, 168, ""},
	//	{"生产", 16.0, "V100", 16, 128, "DeepSeek-R1-671B", 30, 96, ""},
	//	{"生产", 16.0, "V100", 16, 128, "Qwen2.5-v1-72B", 52.5, 168, ""},
	//	{"生产", 16.0, "V100", 16, 128, "空置（云借用）", 5, 16, ""},
	//	{"生产", 16.0, "V100", 16, 128, "Qwen2-instruct-72B", 1, 8, ""},
	//}

	f := excelize.NewFile()
	sheetName := "算力分布"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		log.Fatal(err)
	}
	f.SetActiveSheet(index)

	h.writeHeader(f, sheetName)

	// 环境分组信息
	type envGroup struct {
		name  string
		start int
		end   int
	}

	// 型号分组信息
	type modelGroup struct {
		name  string
		start int
		end   int
	}

	// 存储所有环境分组
	var envGroups []envGroup
	currentEnv := ""
	envStart := 0

	// 存储所有型号分组
	var modelGroups []modelGroup
	currentModel := ""
	modelStart := 0

	// 写入数据行
	for rowIndex, row := range newData {
		rowNum := rowIndex + 3 // 数据从第3行开始

		// 环境分组逻辑
		if row.Env != currentEnv {
			// 上一个环境组结束
			if currentEnv != "" && envStart > 0 {
				envGroups = append(envGroups, envGroup{
					name:  currentEnv,
					start: envStart,
					end:   rowNum - 1,
				})
			}
			// 开始新的环境组
			currentEnv = row.Env
			envStart = rowNum
		}

		// 型号分组逻辑（只在相同环境内分组）
		if row.Model != currentModel || row.Env != currentEnv {
			// 上一个型号组结束
			if currentModel != "" && modelStart > 0 {
				modelGroups = append(modelGroups, modelGroup{
					name:  currentModel,
					start: modelStart,
					end:   rowNum - 1,
				})
			}
			// 开始新的型号组
			currentModel = row.Model
			modelStart = rowNum
		}

		// 写入行数据
		h.writeDataRow(f, sheetName, rowNum, row)
	}

	// 添加最后一个环境组
	if currentEnv != "" && envStart > 0 {
		envGroups = append(envGroups, envGroup{
			name:  currentEnv,
			start: envStart,
			end:   len(newData) + 2,
		})
	}

	// 添加最后一个型号组
	if currentModel != "" && modelStart > 0 {
		modelGroups = append(modelGroups, modelGroup{
			name:  currentModel,
			start: modelStart,
			end:   len(newData) + 2,
		})
	}

	// 合并环境单元格（根据环境分组）
	for _, group := range envGroups {
		// 只有当组内有多行时才需要合并
		if group.start < group.end {
			startCell := fmt.Sprintf("A%d", group.start)
			endCell := fmt.Sprintf("A%d", group.end)
			if err := f.MergeCell(sheetName, startCell, endCell); err != nil {
				fmt.Printf("合并环境单元格 %s:%s 失败: %v\n", startCell, endCell, err)
			}
		}
	}

	// 合并型号和算力单元格（根据型号分组）
	for _, group := range modelGroups {
		// 只有当组内有多行时才需要合并
		if group.start < group.end {
			// 合并型号列
			startCell := fmt.Sprintf("C%d", group.start)
			endCell := fmt.Sprintf("C%d", group.end)
			if err := f.MergeCell(sheetName, startCell, endCell); err != nil {
				fmt.Printf("合并型号单元格 %s:%s 失败: %v\n", startCell, endCell, err)
			}

			// 合并算力列（第一个算力列）
			startCell = fmt.Sprintf("B%d", group.start)
			endCell = fmt.Sprintf("B%d", group.end)
			if err := f.MergeCell(sheetName, startCell, endCell); err != nil {
				fmt.Printf("合并算力单元格 %s:%s 失败: %v\n", startCell, endCell, err)
			}
		}
	}
	// 5. 将 Excel 内容写入缓冲区
	timeStr := util.GetTimeMinite()
	fileName := "新算力精确分布表" + timeStr + ".xlsx"
	fmt.Println("Excel文件生成成功: 新算力精确分布表.xlsx")
	filePath := fmt.Sprintf("./files/%s", fileName)
	if err := f.SaveAs(filePath); err != nil {
		log.Fatal(err)
		return ""
	}
	return fileName

	//return buf
}

func (h *HighLevel) writeHeader(f *excelize.File, sheet string) {
	// 第一行表头
	f.SetCellValue(sheet, "A1", "算力分布情况")
	f.MergeCell(sheet, "A1", "J1")
	styleHeader1, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Font:      &excelize.Font{Bold: true, Size: 14},
	})
	f.SetCellStyle(sheet, "A1", "J1", styleHeader1)

	headerStyle := h.createHeaderStyle(f)
	f.SetCellStyle(sheet, "A1", "J1", headerStyle)

	// 第二行表头
	colNames := []string{"环境", "算力(P)", "型号", "服务器数(台)", "卡数(张)", "算力(P)", "模型", "使用算力", "使用卡数", "备注"}
	for colIdx, name := range colNames {
		col := string(rune('A'+colIdx)) + "2"
		f.SetCellValue(sheet, col, name)
	}

	styleHeader2, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Font:      &excelize.Font{Bold: true},
	})
	f.SetCellStyle(sheet, "A2", "J2", styleHeader2)
	f.SetCellStyle(sheet, "A2", "J2", headerStyle)
}

func (h *HighLevel) writeDataRow(f *excelize.File, sheet string, rowNum int, data DataRow) {
	rowStr := strconv.Itoa(rowNum)
	values := []interface{}{
		data.Env,
		data.ComputeP,
		data.Model,
		data.ServerNum,
		data.CardNum,
		data.ModelUsed,
		data.UsedCompute,
		data.UsedCard,
		data.Remarks,
	}

	style, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	for colIdx, val := range values {
		cell := string(rune('A'+colIdx)) + rowStr
		f.SetCellValue(sheet, cell, val)
		f.SetCellStyle(sheet, cell, cell, style)
	}
}

func (h *HighLevel) mergeCells(f *excelize.File, sheet string, data []DataRow) {
	if len(data) == 0 {
		return
	}

	// 环境列合并
	currentEnv := data[0].Env
	envStartRow := 3
	for i := 1; i <= len(data); i++ {
		if i == len(data) || data[i].Env != currentEnv {
			endRow := i + 2
			if envStartRow < endRow-1 {
				f.MergeCell(sheet, "A"+strconv.Itoa(envStartRow), "A"+strconv.Itoa(endRow-1))
			}
			if i < len(data) {
				currentEnv = data[i].Env
				envStartRow = endRow
			}
		}
	}

	// 算力列和型号列合并（按型号分组）
	currentModel := data[0].Model
	modelStartRow := 3
	computeStartRow := 3 // 算力列与型号列同步合并

	for i := 1; i <= len(data); i++ {
		// 环境发生变化时重置
		if i < len(data) && data[i].Env != data[i-1].Env {
			// 处理前一个环境的最后一组型号
			if modelStartRow < i+2-1 {
				f.MergeCell(sheet, "C"+strconv.Itoa(modelStartRow), "C"+strconv.Itoa(i+2-1))
				f.MergeCell(sheet, "B"+strconv.Itoa(computeStartRow), "B"+strconv.Itoa(i+2-1))
			}

			// 重新开始新的环境
			currentModel = data[i].Model
			modelStartRow = i + 2
			computeStartRow = i + 2
			continue
		}

		if i == len(data) || data[i].Model != currentModel { //3,4    6,6
			endRow := i + 2
			// 合并型号列

			fmt.Println(modelStartRow)
			fmt.Println(endRow)

			f.MergeCell(sheet, "C"+strconv.Itoa(modelStartRow), "C"+strconv.Itoa(endRow))
			// 合并算力列
			f.MergeCell(sheet, "B"+strconv.Itoa(computeStartRow), "B"+strconv.Itoa(endRow))

			modelStartRow = endRow
			computeStartRow = endRow

			if i < len(data) {
				currentModel = data[i].Model

			}
		}
	}
}

func (h *HighLevel) createHeaderStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1}, {Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1}, {Type: "right", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	return style
}
