package parser

import (
	"fmt"
	"github.com/chinamobile/nlpt/pkg/go-restful"
	"github.com/tealeg/xlsx"
	"io"
	"k8s.io/klog"
	"os"
	"path/filepath"
)

type ParseResponse struct {
	ParseData [][]string
}

func ParseApisFromExcel(req *restful.Request,response *restful.Response,spec *ApiExcelSpec)(ParseResponse,error){
	if err :=req.Request.ParseMultipartForm(32<<20);err !=nil{
		return ParseResponse{},fmt.Errorf("failed to import apis:%+v", err)
	}

	file,handle,err :=req.Request.FormFile(spec.MultiPartFileKey)
	klog.Infof("File name: %+v", handle.Filename)
	if err!=nil{
		klog.Error("file err.")
		return ParseResponse{}, fmt.Errorf("invalid file format:%+v", err)
	}

	defer file.Close()

	f,err:=os.OpenFile("./"+handle.Filename,os.O_RDWR|os.O_CREATE,0666)
	if _, err:=io.Copy(f,file);err !=nil{
		klog.Error("import failed,copy file error.")
		return ParseResponse{}, err
	}

	//获取已拷贝文件的绝对路径
	fp , err:=filepath.Abs(filepath.Dir(f.Name()))
	if err != nil {
		return ParseResponse{}, fmt.Errorf("failed to import apis:%+v", err)
	}
	fp = fp +"/" + handle.Filename

	defer os.Remove(fp)

 	excelf,err:=xlsx.OpenFile(fp)
	if err != nil {
		return ParseResponse{}, fmt.Errorf("import failed:%+v", err)
	}

	var apis [][]string
	for _, sheet := range excelf.Sheets {
		if sheet.Name==spec.SheetName{
			for index, field := range spec.TitleRowSpecList {
				cell := sheet.Cell(0, index)

				if len(cell.Value) == 0 {
					klog.Error("invalid file format.")
					return ParseResponse{}, fmt.Errorf("invalid file format")
				}

				if cell.Value != field {
					klog.Error("invalid file format.")
					return ParseResponse{}, fmt.Errorf("invalid file format")
				}
			}
			for i:=1; i<len(sheet.Rows); i++ {
					var api []string
				for j := 0; j < len(spec.TitleRowSpecList); j++ {
					cell := sheet.Cell(i, j)
					klog.Info(cell.Value)
					if len(cell.Value) == 0 {
						continue
					}
					api = append(api, cell.Value)

				}
				apis = append(apis, api)
			}
		}
	}

	return ParseResponse{
		ParseData: apis,
	}, nil
}