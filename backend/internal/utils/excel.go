package utils

import (
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// ExcelRow Excel行数据
type ExcelRow struct {
	BizID    string // 业务流水号
	Amount   string // 金额
	Sender   string // 付款方
	Receiver string // 收款方
	TxType   int8   // 交易类型
}

// ParseExcelFile 解析Excel文件
// 支持的格式: .xlsx, .xls
// Excel格式要求:
//   - 第一行为表头
//   - 必须包含列: 业务流水号, 金额, 付款方, 收款方
//   - 可选列: 交易类型
func ParseExcelFile(filePath string) ([]ExcelRow, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	// 获取第一个工作表
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheet found in excel file")
	}

	// 读取所有行
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) <= 1 {
		return nil, fmt.Errorf("excel file is empty or only has header")
	}

	// 解析表头,获取列索引
	header := rows[0]
	colIndexMap, err := parseHeader(header)
	if err != nil {
		return nil, err
	}

	// 解析数据行
	var excelRows []ExcelRow
	for i := 1; i < len(rows); i++ {
		row := rows[i]

		// 跳过空行
		if isEmptyRow(row) {
			continue
		}

		excelRow, err := parseRow(row, colIndexMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse row %d: %w", i+1, err)
		}

		excelRows = append(excelRows, excelRow)
	}

	return excelRows, nil
}

// parseHeader 解析表头,返回列名到索引的映射
func parseHeader(header []string) (map[string]int, error) {
	colIndexMap := make(map[string]int)

	requiredColumns := []string{"业务流水号", "金额", "付款方", "收款方"}
	foundColumns := make(map[string]bool)

	for colIdx, cell := range header {
		colIndexMap[cell] = colIdx

		for _, reqCol := range requiredColumns {
			if cell == reqCol {
				foundColumns[reqCol] = true
			}
		}
	}

	// 检查必需列是否都存在
	for _, reqCol := range requiredColumns {
		if !foundColumns[reqCol] {
			return nil, fmt.Errorf("missing required column: %s", reqCol)
		}
	}

	return colIndexMap, nil
}

// parseRow 解析单行数据
func parseRow(row []string, colIndexMap map[string]int) (ExcelRow, error) {
	var excelRow ExcelRow

	// 业务流水号 (必需)
	bizIDIdx, ok := colIndexMap["业务流水号"]
	if !ok || bizIDIdx >= len(row) {
		return excelRow, fmt.Errorf("invalid biz_id column")
	}
	excelRow.BizID = row[bizIDIdx]
	if excelRow.BizID == "" {
		return excelRow, fmt.Errorf("biz_id is empty")
	}

	// 金额 (必需)
	amountIdx, ok := colIndexMap["金额"]
	if !ok || amountIdx >= len(row) {
		return excelRow, fmt.Errorf("invalid amount column")
	}
	excelRow.Amount = row[amountIdx]
	if excelRow.Amount == "" {
		return excelRow, fmt.Errorf("amount is empty")
	}

	// 付款方 (必需)
	senderIdx, ok := colIndexMap["付款方"]
	if !ok || senderIdx >= len(row) {
		return excelRow, fmt.Errorf("invalid sender column")
	}
	excelRow.Sender = row[senderIdx]
	if excelRow.Sender == "" {
		return excelRow, fmt.Errorf("sender is empty")
	}

	// 收款方 (必需)
	receiverIdx, ok := colIndexMap["收款方"]
	if !ok || receiverIdx >= len(row) {
		return excelRow, fmt.Errorf("invalid receiver column")
	}
	excelRow.Receiver = row[receiverIdx]
	if excelRow.Receiver == "" {
		return excelRow, fmt.Errorf("receiver is empty")
	}

	// 交易类型 (可选,默认为1)
	txTypeIdx, ok := colIndexMap["交易类型"]
	if ok && txTypeIdx < len(row) && row[txTypeIdx] != "" {
		txType, err := strconv.Atoi(row[txTypeIdx])
		if err != nil {
			return excelRow, fmt.Errorf("invalid tx_type: %w", err)
		}
		excelRow.TxType = int8(txType)
	} else {
		excelRow.TxType = 1 // 默认为转账
	}

	return excelRow, nil
}

// isEmptyRow 检查是否为空行
func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if cell != "" {
			return false
		}
	}
	return true
}

// CreateExcelTemplate 创建Excel模板文件
// 用于用户下载填写
func CreateExcelTemplate(filePath string) error {
	f := excelize.NewFile()
	defer f.Close()

	// 设置表头
	headers := []string{"业务流水号", "金额", "付款方", "收款方", "交易类型"}
	sheetName := "Sheet1"

	for colIdx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetName, cell, header)

		// 设置表头样式(可选)
		style, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	// 添加示例数据
	examples := []interface{}{
		"TX20260113001", "1000000", "机构A", "机构B", "1",
		"TX20260113002", "2000000", "机构B", "机构C", "1",
	}

	for colIdx, example := range examples {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 2)
		f.SetCellValue(sheetName, cell, example)
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "E", 20)

	if err := f.SaveAs(filePath); err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	return nil
}
