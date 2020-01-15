package util

import (
	"fmt"
	"strconv"
)

const (
	defaultQueryPage = 1
	defaultQuerySize = 10
)

type PageStruct struct {
	Content   []interface{} `json:"content"`
	Page      int           `json:"page"`
	Size      int           `json:"size"`
	TotalPage int           `json:"totalPage"`
	TotalSize int           `json:"totalSize"`
}

type listable interface {
	Length() int
	GetItem(i int) (interface{}, error)
}

func PageWrap(items listable, pagestr, sizestr string) (*PageStruct, error) {
	var err error
	var page, size int
	if len(pagestr) == 0 {
		page = defaultQueryPage
	} else {
		page, err = strconv.Atoi(pagestr)
	}
	if err != nil {
		return nil, fmt.Errorf("catnot parse page to int: %+v", err)
	}
	if len(sizestr) == 0 {
		size = defaultQuerySize
	} else {
		size, err = strconv.Atoi(sizestr)
	}
	if err != nil {
		return nil, fmt.Errorf("catnot parse size to int: %+v", err)
	}
	if page <= 0 || size <= 0 {
		return nil, fmt.Errorf("page or size not positive")
	}
	offset := (page - 1) * size
	leng := items.Length()
	if leng == 0 {
		return &PageStruct{
			Page:      page,
			Size:      size,
			TotalPage: 0,
			TotalSize: 0,
			Content:   []interface{}{},
		}, nil
	}
	if offset >= leng {
		return nil, fmt.Errorf("page overflow")
	}
	end := offset + size
	if end > leng {
		end = leng
	}
	total := leng / size
	if leng%size != 0 {
		total = total + 1
	}
	content := make([]interface{}, end-offset)
	for i := range content {
		content[i], err = items.GetItem(i + offset)
		if err != nil {
			return nil, fmt.Errorf("get item error: %+v", err)
		}
	}
	return &PageStruct{
		Page:      page,
		Size:      size,
		TotalPage: total,
		TotalSize: leng,
		//Content:   items[offset:end],
		Content: content,
	}, nil
}
