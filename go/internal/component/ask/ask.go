package ask

import (
	"math"
)

// R 基础响应体
type R struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Code    int         `json:"code"`
}

// ListResponse 列表响应体
type ListResponse struct {
	List       interface{} `json:"list"`
	Extend     interface{} `json:"extend,omitempty"`
	TotalPage  int64       `json:"total_page"`
	TotalCount int64       `json:"total_count"`
	Size       int         `json:"size"`
	Page       int         `json:"page"`
}

type AuthorsListResponse struct {
	List       interface{} `json:"list"`
	TotalPage  int64       `json:"total_page"`
	TotalCount int64       `json:"total_count"`
	UsedCount  int64       `json:"used_count"`
	Size       int         `json:"size"`
	Page       int         `json:"page"`
}
type TrendInt64 struct {
	TimeNode int64 `json:"time_node"`
	Value    int64 `json:"value"`
}

// Success 响应成功
func Success(data ...interface{}) *R {
	return Message("success", data...)
}

// Message 响应成功
func Message(message string, data ...interface{}) *R {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	resp := &R{
		Message: message,
		Code:    0,
	}
	if d != nil {
		resp.Data = d
	}
	return resp
}

type IPageQuery interface {
	GetPage() int
	GetSize() int
}

type PageQuery struct {
	Page    int    `json:"page" form:"page,default=1" binding:"required,gte=1"`
	Size    int    `json:"size" form:"size,default=20" binding:"required,lt=10001"`
	Keyword string `json:"keyword" form:"keyword"`
}

func (p *PageQuery) GetPage() int {
	return p.Page
}

func (p *PageQuery) Offset() int {
	return (p.Page - 1) * p.Size
}

func (p *PageQuery) GetSize() int {
	return p.Size
}

func Page(list interface{}, pageQuery IPageQuery, totalCount int64, ext ...interface{}) *ListResponse {
	totalPage := int64(math.Ceil(float64(totalCount) / float64(pageQuery.GetSize())))
	listResponse := &ListResponse{
		List:       list,
		TotalPage:  totalPage,
		TotalCount: totalCount,
		Size:       pageQuery.GetSize(),
		Page:       pageQuery.GetPage(),
	}
	if len(ext) > 0 {
		listResponse.Extend = ext[0]
	}
	return listResponse
}

func AuthorsPage(list interface{}, pageQuery IPageQuery, totalCount int64, usedCount int64) *AuthorsListResponse {
	totalPage := int64(math.Ceil(float64(totalCount) / float64(pageQuery.GetSize())))
	listResponse := &AuthorsListResponse{
		List:       list,
		TotalPage:  totalPage,
		UsedCount:  usedCount,
		TotalCount: totalCount,
		Size:       pageQuery.GetSize(),
		Page:       pageQuery.GetPage(),
	}
	return listResponse
}
