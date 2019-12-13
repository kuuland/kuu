package kuu

import (
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"mime/multipart"
)

// ParseExcelFromFileHeader
func ParseExcelFromFileHeader(fh *multipart.FileHeader, index int, sheetName string) (rows [][]string, err error) {
	var (
		file multipart.File
		f    *excelize.File
	)

	// 解析Excel
	if file, err = fh.Open(); err != nil {
		return rows, err
	}
	defer ERROR(file.Close())
	if f, err = excelize.OpenReader(file); err != nil {
		return rows, err
	}
	// 选择工作表
	if sheetName == "" && index > 0 {
		sheetName = f.GetSheetName(index)
	}
	if sheetName == "" {
		sheetName = f.GetSheetName(f.GetActiveSheetIndex())
	}
	if sheetName == "" {
		sheetName = f.GetSheetName(1)
	}
	// 读取行
	rows, err = f.GetRows(sheetName)
	return
}
