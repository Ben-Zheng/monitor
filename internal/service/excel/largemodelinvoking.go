package excel

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"log"
	"monitor/util"
	"os"
)

// ServiceRecord 定义服务调用记录结构
type ServiceRecord struct {
	Environment       string `json:"environment"`       //环境
	SerialNumber      int    `json:"serialNumber"`      //序号
	Scene             string `json:"scene"`             //场景名称
	Department        string `json:"department"`        //开发部门
	ApplyModel        string `json:"applyModel"`        //应用模式
	ResponsiblePerson string `json:"responsiblePerson"` //负责人
	Frequency         string `json:"frequency"`         //调用方式
	Model             string `json:"model"`             // 调用模型
	Concurrency       int    `json:"concurrency"`       //申请并发
	CallVolume        int    `json:"callVolume"`        //本期调用
}

type LargeInvokingexcel struct {
}

func NewLargeInvokingexcel() *LargeInvokingexcel {
	return &LargeInvokingexcel{}
}

// 智能平台大模型服务调用情况表
func (l *LargeInvokingexcel) GenerateLedgerExcel(newData []ServiceRecord) string {
	// 示例数据 - 替换为实际数据
	//newData = []ServiceRecord{
	//	{"生产", 1, "知识工程平台知识增强与语料标注", "表单识别", "人工智能中心", "程思香", "实时", "DeepSeek-R1-Distill-Qwen-32B", 144, 4730},
	//	{"生产", 2, "知识工程平台知识增强与语料标注", "表单识别", "人工智能中心", "程思香", "实时", "QwQ-32B", 24, 4730},
	//	{"生产", 3, "知识工程平台知识增强与语料标注", "表单识别", "人工智能中心", "程思香", "实时", "Qwen2-instruct-72B", 848, 4729},
	//	{"生产", 4, "运营管理部境外个人贷款收购付订", "表单识别", "人工智能中心", "程思香", "实时", "Qwen3-8B", 168, 5721},
	//	{"生产", 5, "运营智能审录流程-成都分行车贷保单识别", "表单识别", "人工智能中心", "王琴琴、刘璐", "实时", "DeepSeek-R1-671B", 96, 6391},
	//	{"生产", 6, "运营智能审录流程-城外机构帐户开户", "表单识别", "人工智能中心", "翟雅承", "实时", "Qwen2.5-v1-72B", 168, 0},
	//	{"生产", 7, "信用卡审批支持助手", "表单识别", "数据管理部", "吕伟", "实时", "Qwen3-32B", 80, 0},
	//	{"开发", 8, "空置（云借用）", "", "", "", "", "", 16, 4293},
	//	{"开发", 9, "空置（云借用）", "", "", "", "", "Qwen2.5-v1-32B", 27, 0},
	//	{"开发", 10, "贷前投揭进件材料智能记录", "表单识别", "数据管理部", "李霄昊", "实时", "Qwen3-32B", 1, 0},
	//}

	// 创建新的Excel文件
	f := excelize.NewFile()
	sheet := "服务调用情况"
	f.SetSheetName("Sheet1", sheet)

	// 设置主标题（合并A1:J1）
	f.SetCellValue(sheet, "A1", "2025年7月28日-8月1日智能平台大模型服务调用情况表")
	if err := f.MergeCell(sheet, "A1", "J1"); err != nil {
		fmt.Println("合并标题单元格失败:", err)
		os.Exit(1)
	}

	headerStyle := l.createHeaderStyle(f)
	f.SetCellStyle(sheet, "A1", "A1", headerStyle)
	// 设置表头行
	headers := []string{
		"环境", "序号", "场景", "开发部门", "中心", "负责人",
		"调用频度", "调用模型", "申请并发(页)", "本期调用量",
	}

	// 写入表头
	for col, header := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+col, 2)
		f.SetCellValue(sheet, cell, header)
	}
	f.SetCellStyle(sheet, "A2", "J2", headerStyle)

	// 设置数据样式
	dataStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		fmt.Println("创建数据样式失败:", err)
		os.Exit(1)
	}

	// 场景分组信息
	type sceneGroup struct {
		name  string
		start int // 起始行号（从1开始）
		end   int // 结束行号（从1开始）
	}

	// 存储所有场景分组
	var sceneGroups []sceneGroup
	currentScene := ""
	sceneStart := 0

	// 环境分组信息
	type envGroup struct {
		name  string
		start int
		end   int
	}

	// 存储所有环境分组
	var envGroups []envGroup
	currentEnv := ""
	envStart := 0

	// 写入数据行
	for rowIndex, record := range newData {
		rowNum := rowIndex + 3 // 数据从第3行开始

		// 环境分组逻辑
		if record.Environment != currentEnv {
			// 上一个环境组结束
			if currentEnv != "" && envStart > 0 {
				envGroups = append(envGroups, envGroup{
					name:  currentEnv,
					start: envStart,
					end:   rowNum - 1,
				})
			}
			// 开始新的环境组
			currentEnv = record.Environment
			envStart = rowNum
		}

		// 场景分组逻辑
		if record.Scene != currentScene {
			// 上一个场景组结束
			if currentScene != "" && sceneStart > 0 {
				sceneGroups = append(sceneGroups, sceneGroup{
					name:  currentScene,
					start: sceneStart,
					end:   rowNum - 1,
				})
			}
			// 开始新的场景组
			currentScene = record.Scene
			sceneStart = rowNum
		}

		// 写入行数据
		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowNum), record.Environment)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowNum), record.SerialNumber)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowNum), record.Scene)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowNum), record.Department)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", rowNum), record.ResponsiblePerson)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", rowNum), record.Frequency)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", rowNum), record.Model)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", rowNum), record.Concurrency)
		f.SetCellValue(sheet, fmt.Sprintf("J%d", rowNum), record.CallVolume)

		// 应用数据样式
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("J%d", rowNum), dataStyle)
	}

	// 添加最后一个环境组
	if currentEnv != "" && envStart > 0 {
		envGroups = append(envGroups, envGroup{
			name:  currentEnv,
			start: envStart,
			end:   len(newData) + 2,
		})
	}

	// 添加最后一个场景组
	if currentScene != "" && sceneStart > 0 {
		sceneGroups = append(sceneGroups, sceneGroup{
			name:  currentScene,
			start: sceneStart,
			end:   len(newData) + 2,
		})
	}

	// 合并环境单元格（根据环境分组）
	for _, group := range envGroups {
		// 只有当组内有多行时才需要合并
		if group.start < group.end {
			startCell := fmt.Sprintf("A%d", group.start)
			endCell := fmt.Sprintf("A%d", group.end)
			if err := f.MergeCell(sheet, startCell, endCell); err != nil {
				fmt.Printf("合并环境单元格 %s:%s 失败: %v\n", startCell, endCell, err)
			}
		}
	}

	// 合并场景单元格（根据场景分组）
	for _, group := range sceneGroups {
		// 只有当组内有多行时才需要合并
		if group.start < group.end {
			// 合并场景列
			startCell := fmt.Sprintf("C%d", group.start)
			endCell := fmt.Sprintf("C%d", group.end)
			if err := f.MergeCell(sheet, startCell, endCell); err != nil {
				fmt.Printf("合并场景单元格 %s:%s 失败: %v\n", startCell, endCell, err)
			}

			// 合并开发部门列
			startCell = fmt.Sprintf("D%d", group.start)
			endCell = fmt.Sprintf("D%d", group.end)
			if err := f.MergeCell(sheet, startCell, endCell); err != nil {
				fmt.Printf("合并开发部门单元格 %s:%s 失败: %v\n", startCell, endCell, err)
			}

			// 合并中心列
			startCell = fmt.Sprintf("E%d", group.start)
			endCell = fmt.Sprintf("E%d", group.end)
			if err := f.MergeCell(sheet, startCell, endCell); err != nil {
				fmt.Printf("合并中心单元格 %s:%s 失败: %v\n", startCell, endCell, err)
			}

			// 合并负责人列
			startCell = fmt.Sprintf("F%d", group.start)
			endCell = fmt.Sprintf("F%d", group.end)
			if err := f.MergeCell(sheet, startCell, endCell); err != nil {
				fmt.Printf("合并负责人单元格 %s:%s 失败: %v\n", startCell, endCell, err)
			}

			// 合并调用频度列
			startCell = fmt.Sprintf("G%d", group.start)
			endCell = fmt.Sprintf("G%d", group.end)
			if err := f.MergeCell(sheet, startCell, endCell); err != nil {
				fmt.Printf("合并调用频度单元格 %s:%s 失败: %v\n", startCell, endCell, err)
			}
		}
	}

	// 设置列宽（根据实际情况调整）
	f.SetColWidth(sheet, "A", "A", 10) // 环境
	f.SetColWidth(sheet, "B", "B", 6)  // 序号
	f.SetColWidth(sheet, "C", "C", 35) // 场景
	f.SetColWidth(sheet, "D", "D", 15) // 开发部门
	f.SetColWidth(sheet, "E", "E", 20) // 中心
	f.SetColWidth(sheet, "F", "F", 20) // 负责人
	f.SetColWidth(sheet, "G", "G", 10) // 调用频度
	f.SetColWidth(sheet, "H", "H", 30) // 调用模型
	f.SetColWidth(sheet, "I", "I", 15) // 申请并发(页)
	f.SetColWidth(sheet, "J", "J", 15) // 本期调用量

	// 保存文件
	//if err := f.SaveAs("智能平台服务调用情况表.xlsx"); err != nil {
	//	fmt.Println("保存文件失败:", err)
	//	os.Exit(1)
	//}
	timeStr := util.GetTimeMinite()
	fileName := "智能平台大模型服务调用情况表" + timeStr + ".xlsx"
	filePath := fmt.Sprintf("./files/%s", fileName)

	if err := f.SaveAs(filePath); err != nil {
		log.Println(err)
		return fileName
	}
	return ""

}

func (l *LargeInvokingexcel) createHeaderStyle(f *excelize.File) int {
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
