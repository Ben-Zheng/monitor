package task

import (
	"context"
	"fmt"
	"github.com/xuri/excelize/v2"
	"sync"
	"testing"
	"time"
)

func TestTaskExecutionOrder(t *testing.T) {
	scheduler := NewDelayedTaskScheduler()
	scheduler.Start(3) // 单工作线程保证顺序
	defer scheduler.Stop()

	// 记录任务执行顺序
	var executionOrder []string
	var mu sync.Mutex
	// 添加任务（乱序添加）
	go scheduler.Schedule("C", 10000*time.Millisecond, func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		fmt.Println("执行C")
		executionOrder = append(executionOrder, "C")
		return nil
	})

	go scheduler.Schedule("A", 2000*time.Millisecond, func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		fmt.Println("执行A")
		executionOrder = append(executionOrder, "A")
		return nil
	})

	go scheduler.Schedule("B", 12000*time.Millisecond, func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		fmt.Println("执行B")
		executionOrder = append(executionOrder, "B")
		return nil
	})

	go scheduler.Schedule("F", 14000*time.Millisecond, func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		fmt.Println("执行F")
		executionOrder = append(executionOrder, "F")
		return nil
	})

	go scheduler.Schedule("E", 20000*time.Millisecond, func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		fmt.Println("执行E")
		executionOrder = append(executionOrder, "E")
		return nil
	})

	go scheduler.Schedule("G", 8000*time.Millisecond, func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		fmt.Println("执行G")
		executionOrder = append(executionOrder, "G")
		return nil
	})

	// 等待所有任务完成
	for {

	}

}

func TestTaskExecutionExcelNew(t *testing.T) {
	f := excelize.NewFile()
	const sheetName = "算力分布"
	f.NewSheet(sheetName)
	f.SetActiveSheet(1)

	// 单元格样式定义
	titleStyle := createTitleStyle(f)
	headerStyle := createHeaderStyle(f)
	dataStyle := createDataStyle(f)
	groupStyle := createGroupStyle(f)

	// === 表格框架构建 ===
	// 主标题
	f.SetCellValue(sheetName, "A1", "算力分布情况")
	f.MergeCell(sheetName, "A1", "H1")
	f.SetRowHeight(sheetName, 1, 28)
	f.SetCellStyle(sheetName, "A1", "H1", titleStyle)

	// 表头行
	f.SetCellValue(sheetName, "A2", "类型")
	f.SetCellValue(sheetName, "B2", "环境")
	f.SetCellValue(sheetName, "C2", "算力(P)")
	f.SetCellValue(sheetName, "D2", "算力分布")
	f.MergeCell(sheetName, "D2", "G2")
	f.SetCellValue(sheetName, "H2", "模型部署分布")
	f.SetRowHeight(sheetName, 2, 22)
	f.SetCellStyle(sheetName, "A2", "H2", headerStyle)

	// 子表头行
	subHeaders := []string{"型号", "服务器数(台)", "卡数(张)", "算力(P)", "模型", "使用算力", "使用卡数"}
	for i, h := range subHeaders {
		col, _ := excelize.ColumnNumberToName(i + 4) // D列开始
		f.SetCellValue(sheetName, col+"3", h)
	}
	f.SetRowHeight(sheetName, 3, 22)
	f.SetCellStyle(sheetName, "D3", "H3", headerStyle)

	// === 精确数据填充（按图片内容）===
	// 生产环境数据
	createProductionSection(f, sheetName, dataStyle)

	// 开发环境数据
	createDevelopmentSection(f, sheetName, dataStyle)

	// === 脚注与汇总信息 ===
	f.SetCellValue(sheetName, "A26", "1.高性算力从大模型部署情况")
	f.SetCellValue(sheetName, "A27", "2.大模型调用量情况")
	f.SetCellValue(sheetName, "A28", "2.2模型支撑场景调用量情况")
	f.SetCellValue(sheetName, "A29", "Sheeh")
	f.SetCellValue(sheetName, "A30", "2.0大模型")

	// 计算并添加汇总行
	totalRow := 31
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", totalRow), "共计：")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", totalRow), "自购算力：35.6（N）+486.5，共522.1P")
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", totalRow), "其中生产25（N）+436.5，共461.5P")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", totalRow), "开发环境10.6（N）+50，共60.6P")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("H%d", totalRow))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("H%d", totalRow), groupStyle)

	// 列宽设置（精确匹配图片宽度）
	colWidths := []float64{8, 10, 8, 15, 12, 10, 15, 12}
	for i, width := range colWidths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheetName, col, col, width)
	}

	// 保存文件
	if err := f.SaveAs("算力精确分布表.xlsx"); err != nil {
		fmt.Println("保存失败:", err)
	} else {
		fmt.Println("Excel文件已精确生成: 算力精确分布表.xlsx")
	}
}

// 样式创建函数
func createTitleStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	return style
}

func createHeaderStyle(f *excelize.File) int {
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

func createDataStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#BFBFBF", Style: 1}, {Type: "top", Color: "#BFBFBF", Style: 1},
			{Type: "bottom", Color: "#BFBFBF", Style: 1}, {Type: "right", Color: "#BFBFBF", Style: 1},
		},
	})
	return style
}

func createGroupStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D9E1F2"}, Pattern: 1},
	})
	return style
}

// 生产环境部分创建
func createProductionSection(f *excelize.File, sheet string, dataStyle int) {
	// 生产主分组
	f.SetCellValue(sheet, "A4", "生产")
	f.MergeCell(sheet, "A4", "A13") // 合并12行

	// 生产环境分类型
	types := []struct {
		text     string
		startRow int
		endRow   int
	}{
		{"461.5P", 4, 7},
		{"V100", 8, 10},
		{"A800", 11, 13},
	}

	for _, t := range types {
		row := t.startRow
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), t.text)
		f.MergeCell(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", t.endRow))
	}

	// 910B型号数据（4-7行）
	f.SetCellValue(sheet, "C4", "910B")
	f.MergeCell(sheet, "C4", "C7")
	f.SetCellValue(sheet, "D4", "175")
	f.MergeCell(sheet, "D4", "D7")
	f.SetCellValue(sheet, "E4", "1400")
	f.MergeCell(sheet, "E4", "E7")
	f.SetCellValue(sheet, "F4", "436.5P")
	f.MergeCell(sheet, "F4", "F7")

	models := []struct {
		row   int
		model string
		power string
		cards interface{}
	}{
		{4, "DeepSeek-R1-Distill-Qwen-32B", "45P", 144},
		{5, "QwQ-32B", "7.5P", 24},
		{6, "Qwen2-instruct-72B", "264P", 848},
		{7, "", "", ""},
	}

	for _, m := range models {
		f.SetCellValue(sheet, fmt.Sprintf("G%d", m.row), m.model)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", m.row), m.power)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", m.row), m.cards)
	}

	// V100型号数据（8-10行）
	f.SetCellValue(sheet, "C8", "V100")
	f.MergeCell(sheet, "C8", "C10")
	f.SetCellValue(sheet, "D8", "16")
	f.MergeCell(sheet, "D8", "D10")
	f.SetCellValue(sheet, "E8", "128")
	f.MergeCell(sheet, "E8", "E10")
	f.SetCellValue(sheet, "F8", "16P")
	f.MergeCell(sheet, "F8", "F10")

	v100Models := []struct {
		row   int
		model string
		power string
		cards interface{}
	}{
		{8, "DeepSeek-R1-671B", "30P", 96},
		{9, "Qwen2.5-v1-72B", "52.5P", 168},
		{10, "Qwen3-32B", "25P", 80},
	}

	for _, m := range v100Models {
		f.SetCellValue(sheet, fmt.Sprintf("G%d", m.row), m.model)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", m.row), m.power)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", m.row), m.cards)
	}

	// A800型号数据
	f.SetCellValue(sheet, "C11", "A800")
	f.SetCellValue(sheet, "D11", "14")
	f.SetCellValue(sheet, "E11", "28")
	f.SetCellValue(sheet, "F11", "9P")
	f.SetCellValue(sheet, "G11", "空置（云借用）")
	f.SetCellValue(sheet, "H11", "10P")
	f.SetCellValue(sheet, "I11", "16P") // 注意：这里需要128张卡？但图片中是16P

	// 设置合并和数据样式
	for row := 4; row <= 13; row++ {
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("I%d", row), dataStyle)
	}
}

// 开发环境部分创建
func createDevelopmentSection(f *excelize.File, sheet string, dataStyle int) {
	startRow := 14
	f.SetCellValue(sheet, fmt.Sprintf("A%d", startRow), "开发")
	f.MergeCell(sheet, fmt.Sprintf("A%d", startRow), fmt.Sprintf("A%d", startRow+16))

	devModels := []struct {
		row   int
		model string
		power string
		cards interface{}
	}{
		{14, "Qwen2.5-v1-32B", "8.7P", 27},
		{15, "Qwen3-32B", "0.3P", 1},
		{16, "大模型服务平台测试", "5P", 16},
		{17, "Qwen2.5-v1-32B", "2.5P", 8},
		{18, "DeepSeek-R1-671B", "5P", 16},
		{19, "Qwen2.5-v1-72B", "2.5P", 8},
		{20, "华为大EP方案测试", "20P", 64},
		{21, "Qwen3-32B（云借用）", "5P", 16},
		{22, "空置（云借用）", "5P", 16},
		{23, "Qwen2-instruct-72B", "1P", 8},
		{24, "大模型蒸馏验证", "1P", 8},
		{25, "Qwen2.5-v1-32B", "0.5P", 4},
		{26, "Qwen2.5-v1-72B", "1P", 8},
		{27, "Qwen3-8B", "0.5P", 4},
		{28, "Qwen3-32B", "1P", 8},
		{29, "中小模型训练", "4P", 32},
		{30, "产品处大模型部署", "1P", 8},
		{31, "DeepSeek-R1-Distill-Qwen-32B（创新借用）", "0.6P", 2},
	}

	for _, m := range devModels {
		row := startRow + m.row - 14
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), m.model)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), m.power)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), m.cards)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("I%d", row), dataStyle)
	}
}
