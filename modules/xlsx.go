package modules

import (
	"mime/multipart"
	"strings"

	"github.com/tealeg/xlsx"
)

type Xlsx struct {
	File    multipart.File
	Handler *multipart.FileHeader
}

func (x *Xlsx) Read() (result []Data, err error) {

	excelFile, err := xlsx.OpenReaderAt(x.File, x.Handler.Size)
	if err != nil {
		return
	}

	for _, sheet := range excelFile.Sheets {
		rows := sheet.Rows[1:] // skip header row
		for _, row := range rows {
			col0 := strings.Trim(row.Cells[0].Value, " ")
			col1 := strings.Trim(row.Cells[1].Value, " ")
			col2 := strings.Trim(row.Cells[2].Value, " ")
			col3 := strings.Trim(row.Cells[3].Value, " ")

			if col0 != "" && col1 != "" && col2 != "" && col3 != "" {
				result = append(result, Data{
					Source: Directory{
						Path: col1,
						Name: col0,
					},
					Destination: Directory{
						Path: col3,
						Name: col2,
					},
				})
			} else if col0 == "" && col1 == "" && col2 == "" && col3 == "" {
				return
			}
		}
	}
	return
}
