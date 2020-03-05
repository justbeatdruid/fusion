package parser

import (
	"errors"
	"github.com/emicklei/go-restful"
	"github.com/tealeg/xlsx"
	"io"
	"k8s.io/klog"
	"os"
	"path/filepath"
)

type ParseResponse struct {
	ParseData [][]string
}

func ParseTopicsFromExcel(req *restful.Request, response *restful.Response, spec *TopicExcelSpec) (ParseResponse, error) {
	req.Request.ParseMultipartForm(32 << 20)
	file, handler, err := req.Request.FormFile(spec.MultiPartFileKey)

	klog.Infof("File name: %+v", handler.Filename)
	if err != nil {
		klog.Error("File error.")
		return ParseResponse{}, errors.New("File error.")
	}

	defer file.Close()

	f, err := os.OpenFile("./"+handler.Filename, os.O_RDWR|os.O_CREATE, 0666)
	if _, err = io.Copy(f, file); err != nil {
		klog.Error("import failed,copy file error.")
		return ParseResponse{}, err
	}

	//获取已拷贝文件的绝对路径
	fp, err := filepath.Abs(filepath.Dir(f.Name()))
	fp = fp + "/" + handler.Filename

	excelf, err := xlsx.OpenFile(fp)
	if err != nil {
		return ParseResponse{}, errors.New("import failed, not excel file")
	}

	var tps [][]string
	for _, sheet := range excelf.Sheets {
		if sheet.Name == spec.SheetName {

			for rNum, _ := range sheet.Rows {
				var tp []string
				for index, field := range spec.TopicExcelDefinitionList {
					cell := sheet.Cell(rNum, index+1)
					klog.Info(cell.Value)
					if len(cell.Value) == 0 {
						klog.Error("invalid file format.")
						return ParseResponse{}, errors.New("invalid file format.")
					}

					if rNum == 1 {
						if cell.Value != field {
							klog.Error("invalid file format.")
							return ParseResponse{}, errors.New("invalid file format.")
						}
					} else {
						tp = append(tp, cell.Value)
					}
				}

				tps = append(tps, tp)
			}

		}
	}

	return ParseResponse{
		ParseData: tps,
	}, nil
}
