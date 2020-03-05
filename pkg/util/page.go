package util

import (
	"fmt"
	"sort"
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
	Len() int
	GetItem(i int) (interface{}, error)
}

type sortable interface {
	listable
	Less(i, j int) bool
	Swap(i, j int)
}

func PageWrap(items listable, pagestr, sizestr string) (*PageStruct, error) {
	// parse parameters
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

	// if need sort
	if sitems, ok := items.(sortable); ok {
		sort.Sort(sitems)
		items = sitems.(listable)
	}

	// select items
	// 1st case: select all
	if page < 0 || size < 0 {
		l := items.Len()
		content := make([]interface{}, l)
		for i := 0; i < l; i++ {
			content[i], err = items.GetItem(i)
			if err != nil {
				return nil, fmt.Errorf("get item error: %+v", err)
			}
		}
		return &PageStruct{
			Page:      1,
			Size:      l,
			TotalPage: 1,
			TotalSize: l,
			Content:   content,
		}, nil
	}
	offset := (page - 1) * size
	// 2nd case: select pages
	leng := items.Len()
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
