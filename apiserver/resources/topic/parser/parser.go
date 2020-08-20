package parser

import (
	"fmt"
	"github.com/chinamobile/nlpt/pkg/go-restful"
	"github.com/tealeg/xlsx"
	"io"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"strings"
)

type ParseResponse struct {
	ParseData [][]string
}

func ParseTopicsFromExcel(req *restful.Request, response *restful.Response, spec *TopicExcelSpec) (ParseResponse, error, string) {

	if err := req.Request.ParseMultipartForm(32 << 20); err != nil {
		return ParseResponse{}, fmt.Errorf("failed to import topics:%+v", err), "读取上传文件失败"
	}
	file, handler, err := req.Request.FormFile(spec.MultiPartFileKey)

	klog.Infof("File name: %+v", handler.Filename)
	if err != nil {
		klog.Error("File error.")
		return ParseResponse{}, fmt.Errorf("invalid file format:%+v", err), "读取上传文件失败"
	}

	if !strings.HasSuffix(handler.Filename, ".xlsx") {
		return ParseResponse{}, fmt.Errorf("只支持上传xlsx后缀的文件"), "只支持上传xlsx后缀的文件"
	}

	defer file.Close()

	f, err := os.OpenFile("./"+handler.Filename, os.O_RDWR|os.O_CREATE, 0666)
	if _, err = io.Copy(f, file); err != nil {
		klog.Error("import failed,copy file error.")
		return ParseResponse{}, err, "读取上传文件失败"
	}

	//获取已拷贝文件的绝对路径
	fp, err := filepath.Abs(filepath.Dir(f.Name()))
	if err != nil {
		return ParseResponse{}, fmt.Errorf("failed to import topics:%+v", err), "读取上传文件失败"
	}
	fp = fp + "/" + handler.Filename

	defer os.Remove(fp)

	excelf, err := xlsx.OpenFile(fp)
	if err != nil {
		return ParseResponse{}, fmt.Errorf("import failed:%+v", err), "读取上传文件失败"
	}

	var tps [][]string
	for _, sheet := range excelf.Sheets {
		if sheet.Name == spec.SheetName {
			for index, field := range spec.TitleRowSpecList {
				cell := sheet.Cell(0, index)

				if len(cell.Value) == 0 {
					klog.Error("invalid file format.")
					return ParseResponse{}, fmt.Errorf("invalid file format"), "文件解析失败，文件格式错误"
				}

				if cell.Value != field {
					klog.Error("invalid file format.")
					return ParseResponse{}, fmt.Errorf("invalid file format"), "文件解析失败，文件格式错误"
				}
			}
			for i := 1; i < len(sheet.Rows); i++ {
				var tp []string
				for j := 0; j < len(spec.TitleRowSpecList); j++ {
					cell := sheet.Cell(i, j)
					klog.Info(cell.Value)
					if len(cell.Value) == 0 {
						continue
					}
					tp = append(tp, cell.Value)

				}

				tps = append(tps, tp)
			}

		}
	}

	return ParseResponse{
		ParseData: tps,
	}, nil, ""
}
