package excel

import (
	"fmt"
	"log"
	"monitor/util"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Record struct {
	Environment string
	Model       string
	Scenario    string
	Department  string
	Center      string
	Manager     string
	Concurrency int
	Success     int
}

type ServiceLedgerDetail struct {
}

func NewServiceLedgerDetail() *ServiceLedgerDetail {
	return &ServiceLedgerDetail{}
}

// 智能平台服务调用情况表
func (s *ServiceLedgerDetail) GenerateServiceLedger(data []Record) string {
	//data = []Record{
	//	// 生产环境 - DeepSeek-R1
	//	{"生产", "DeepSeek-R1-Distill-Qwen-32B", "知识工程平台知识增强与语料标注", "表单识别", "人工智能中心", "程思香", 144, 4730},
	//	{"生产", "DeepSeek-R1-Distill-Qwen-32B", "OA系统公文校对", "表单识别", "人工智能中心", "程思香", 24, 4730},
	//	{"生产", "DeepSeek-R1-Distill-Qwen-32B", "罗盘-知识问答", "表单识别", "人工智能中心", "程思香", 25, 4731},
	//
	//	// 生产环境 - Qwen3-8B
	//	{"生产", "Qwen3-8B", "运营管理部境外个人贷款收购付订", "表单识别", "人工智能中心", "程思香", 168, 5721},
	//	{"生产", "Qwen3-8B", "运营智能审录流程-成都分行车贷保单识别", "表单识别", "人工智能中心", "王琴琴、刘璐", 96, 6391},
	//
	//	// 生产环境 - Qwen2.5-v1-72B
	//	{"生产", "Qwen2.5-v1-72B", "运营智能审录流程-城外机构帐户开户", "表单识别", "人工智能中心", "翟雅承", 168, 0},
	//	{"生产", "Qwen2.5-v1-72B", "信用卡审批支持助手", "表单识别", "数据管理部", "吕伟", 80, 0},
	//
	//	// 开发环境（测试环境）
	//	{"开发", "Qwen2.5-v1-31B", "罗盘-知识问答", "表单识别", "数据管理部", "吕伟", 16, 4293},
	//	{"开发", "Qwen2.5-v1-31B", "罗盘-知识问答", "表单识别", "数据管理部", "吕伟", 27, 0},
	//	{"开发", "Qwen3-32B", "贷前投揭进件材料智能记录", "表单识别", "数据管理部", "李霄昊", 1, 0},
	//}

	// 创建新的Excel文件
	f := excelize.NewFile()
	sheetName := "服务调用情况"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		log.Println(err)
		return ""
	}
	f.SetActiveSheet(index)

	// ================= 1. 添加主标题 =================
	mainTitle := "智能平台服务调用情况表"
	f.MergeCell(sheetName, "A1", "H1")
	f.SetCellValue(sheetName, "A1", mainTitle)
	mainTitleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: getBorderStyle(),
	})
	f.SetCellStyle(sheetName, "A1", "H1", mainTitleStyle)

	// ================= 2. 添加表头 =================
	rowOffset := 1
	headers := []string{"环境", "模型", "场景", "开发部门", "中心", "负责人", "申请并发(页)", "本期成功调用量"}
	for col, header := range headers {
		f.SetCellValue(sheetName, s.getCell(col, 1+rowOffset), header)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF", // 白色
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4472C4"}, // 深蓝色背景
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: getBorderStyle(),
	})
	f.SetCellStyle(sheetName, s.getCell(0, 1+rowOffset), s.getCell(len(headers)-1, 1+rowOffset), headerStyle)

	// ================= 3. 处理数据行 =================
	currentEnv := ""
	currentModel := ""
	envStartRow := make(map[string]int)
	envEndRow := make(map[string]int)
	modelStartRow := make(map[string]int)
	modelEndRow := make(map[string]int)

	// 当前行号（表头之后开始）
	currentRow := 1 + rowOffset
	var mergeRanges []string

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: getBorderStyle(),
	})

	for _, record := range data {
		currentRow++

		// 处理环境变更
		if record.Environment != currentEnv {
			if currentEnv != "" {
				envEndRow[currentEnv] = currentRow - 1
			}
			currentEnv = record.Environment
			envStartRow[currentEnv] = currentRow
		}

		// 处理模型变更
		if record.Model != currentModel {
			if currentModel != "" {
				// 为上一个模型添加合计行
				s.addTotalRow(f, sheetName, currentRow, modelStartRow[currentModel], currentRow-1)
				modelEndRow[currentModel] = currentRow
				currentRow++ // 增加了合计行，再进一行
			}
			currentModel = record.Model
			modelStartRow[currentModel] = currentRow
		}

		// 写入数据行
		f.SetCellValue(sheetName, s.getCell(0, currentRow), record.Environment)
		f.SetCellValue(sheetName, s.getCell(1, currentRow), record.Model)
		f.SetCellValue(sheetName, s.getCell(2, currentRow), record.Scenario)
		f.SetCellValue(sheetName, s.getCell(3, currentRow), record.Department)
		f.SetCellValue(sheetName, s.getCell(4, currentRow), record.Center)
		f.SetCellValue(sheetName, s.getCell(5, currentRow), record.Manager)
		f.SetCellValue(sheetName, s.getCell(6, currentRow), record.Concurrency)
		f.SetCellValue(sheetName, s.getCell(7, currentRow), record.Success)

		// 应用数据行样式（居中，边框）
		f.SetCellStyle(sheetName, s.getCell(0, currentRow), s.getCell(7, currentRow), dataStyle)
	}

	// ================= 4. 处理最后一个模型和环境 =================
	if currentModel != "" {
		s.addTotalRow(f, sheetName, currentRow+1, modelStartRow[currentModel], currentRow)
		modelEndRow[currentModel] = currentRow + 1
		currentRow++ // 增加总计行
	}

	if currentEnv != "" {
		envEndRow[currentEnv] = currentRow
	}

	// ================= 5. 合并单元格 =================
	// 合并环境单元格（A列）
	for env, start := range envStartRow {
		end := envEndRow[env]
		if start < end {
			startCell := s.getCell(0, start)
			endCell := s.getCell(0, end)
			mergeRanges = append(mergeRanges, fmt.Sprintf("%s:%s", startCell, endCell))

			// 设置垂直居中样式
			centerStyle, _ := f.NewStyle(&excelize.Style{
				Alignment: &excelize.Alignment{
					Vertical: "center",
				},
				Border: getBorderStyle(),
			})
			f.SetCellStyle(sheetName, startCell, endCell, centerStyle)
		}
	}

	// 合并模型单元格（B列）
	for model, start := range modelStartRow {
		end := modelEndRow[model]
		if start < end {
			startCell := s.getCell(1, start)
			endCell := s.getCell(1, end)
			mergeRanges = append(mergeRanges, fmt.Sprintf("%s:%s", startCell, endCell))

			// 设置垂直居中样式
			centerStyle, _ := f.NewStyle(&excelize.Style{
				Alignment: &excelize.Alignment{
					Vertical: "center",
				},
				Border: getBorderStyle(),
			})
			f.SetCellStyle(sheetName, startCell, endCell, centerStyle)
		}
	}

	// 应用所有合并
	for _, mergeRange := range mergeRanges {
		parts := strings.Split(mergeRange, ":")
		if len(parts) == 2 {
			f.MergeCell(sheetName, parts[0], parts[1])
		}
	}

	// ================= 6. 设置列宽 =================

	for col := 0; col < len(headers); col++ {
		width := 15.0
		if col == 1 || col == 2 { // 模型和场景列更宽
			width = 35.0
		}
		colName, _ := excelize.ColumnNumberToName(col + 1)
		f.SetColWidth(sheetName, colName, colName, width)
	}

	// ================= 7. 保存文件 =================
	timeStr := util.GetTimeMinite()
	fileName := "智能平台服务调用情况表" + timeStr
	filePath := fmt.Sprintf("./flies/%s.xlsx", fileName)

	if err := f.SaveAs(filePath); err != nil {
		log.Println(err)
		return filePath
	}
	return ""
}

//}
//
//log.Println("Excel文件生成成功：智能平台服务调用情况表.xlsx")
//	buf, err := f.WriteToBuffer()
//	return buf
//}

// 添加合计行（合并场景到负责人列）
func (s *ServiceLedgerDetail) addTotalRow(f *excelize.File, sheetName string, row int, startRow, endRow int) {
	// 计算合计值
	totalConcurrency := 0
	totalSuccess := 0

	// 遍历需要计算的数据行
	for r := startRow; r <= endRow; r++ {
		// 直接获取并发数
		concurrencyCell := s.getCell(6, r)
		concurrency, err := f.GetCellValue(sheetName, concurrencyCell)
		if err != nil {
			continue
		}
		concurrencyInt, err := strconv.Atoi(concurrency)
		if err == nil {
			totalConcurrency += concurrencyInt
		}

		// 直接获取成功数
		successCell := s.getCell(7, r)
		success, err := f.GetCellValue(sheetName, successCell)
		if err != nil {
			continue
		}
		successInt, err := strconv.Atoi(success)
		if err == nil {
			totalSuccess += successInt
		}
	}

	// 写入"总计"到场景列
	f.SetCellValue(sheetName, s.getCell(2, row), "总计")

	// 合并场景到负责人列（C列到F列）
	startMerge := s.getCell(2, row) // C列
	endMerge := s.getCell(5, row)   // F列
	f.MergeCell(sheetName, startMerge, endMerge)

	// 设置总计行样式（浅蓝色背景，加粗，居中）
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D6DCE4"}, // 浅蓝色背景
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: getBorderStyle(),
	})

	// 应用样式到总计行（从环境列到成功调用量列）
	f.SetCellStyle(sheetName, s.getCell(0, row), s.getCell(7, row), totalStyle)

	// 写入数值
	f.SetCellValue(sheetName, s.getCell(6, row), totalConcurrency)
	f.SetCellValue(sheetName, s.getCell(7, row), totalSuccess)
}

// 根据列索引和行号获取单元格名称
func (s *ServiceLedgerDetail) getCell(colIndex, row int) string {
	cell, _ := excelize.CoordinatesToCellName(colIndex+1, row)
	return cell
}

// 获取边框样式
func getBorderStyle() []excelize.Border {
	return []excelize.Border{
		{Type: "left", Color: "000000", Style: 1},
		{Type: "top", Color: "000000", Style: 1},
		{Type: "bottom", Color: "000000", Style: 1},
		{Type: "right", Color: "000000", Style: 1},
	}
}
