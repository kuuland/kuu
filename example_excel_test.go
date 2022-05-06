package kuu

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func ExampleSetData() {
	headers := Headers{
		{
			Label: "字段0",
			Field: "Field0",
			Index: 0,
		},
		{
			Label: "字段3",
			Field: "Field3",
			Index: 3,
		},
		{
			Label: "字段1",
			Field: "Field1",
			Index: 1,
		},
		{
			Label: "字段2",
			Field: "Field2",
			Index: 2,
		},
	}
	list := []map[string]interface{}{
		{
			"Field0": "221",
			"Field1": "221",
			"Field2": "221",
			"Field3": "221",
		},
		{
			"Field0": "2211",
			"Field1": "221",
			"Field2": "221",
			"Field3": "221",
		},
		{
			"Field0": "2212",
			"Field1": "221",
			"Field2": "221",
			"Field3": "221",
		},
		{
			"Field0": "2213",
			"Field1": "221",
			"Field2": "221",
			"Field3": "221",
		},
		{
			"Field0": "2214",
			"Field1": "221",
			"Field2": "221",
			"Field3": "221",
		},
		{
			"Field0": "2215",
			"Field1": "221",
			"Field2": "221",
			"Field3": "221",
		},
		{
			"Field0": "2216",
			"Field1": "221",
			"Field2": "221",
			"Field3": "221",
		},
	}
	file2 := excelize.NewFile()
	SetData(file2, "Sheet1", headers, list)
	file2.SaveAs("data.xlsx")
}

func ExampleReadData() {
	file, err := excelize.OpenFile("data.xlsx")
	if err != nil {
		panic(err)
	}
	headers := Headers{
		{
			Label: "字段0",
			Field: "Field0",
			Index: 0,
		},
		{
			Label: "字段3",
			Field: "Field3",
			Index: 3,
		},
		{
			Label: "字段1",
			Field: "Field1",
			Index: 1,
		},
		{
			Label: "字段2",
			Field: "Field2",
			Index: 2,
		},
	}
	sheet := file.GetSheetName(1)
	list, err := ReadData(file, sheet, headers)
	if err != nil {
		panic(err)
	}
	for _, m := range list {
		fmt.Println(m)
	}
}

func ExampleSetHeader() {
	var headers = Headers{}
	for i := 0; i < 500; i++ {
		h := Header{
			Label: fmt.Sprintf("Label%d", i),
			Field: fmt.Sprintf("Field%d", i),
			Index: i,
		}
		v := h.GetExcelCol()
		fmt.Println(v)
		headers = append(headers, h)
	}
	file := excelize.NewFile()
	SetHeader(file, "Sheet1", headers)
	file.SaveAs("500-col-headers.xlsx")
}
