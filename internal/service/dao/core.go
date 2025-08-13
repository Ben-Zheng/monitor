package dao

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/oklog/ulid"
	"gorm.io/gorm"
)

type OrderParam struct {
	Column string `json:"column" command:"排序字段"`                 // 排序字段
	Order  string `json:"order" command:"排序方式" enums:"asc,desc"` // 排序方式
}

type PageParam struct {
	PageNum  int `json:"pageNum" command:"pageNum" query:"pageNum"`    // 页码
	PageSize int `json:"pageSize" command:"pageSize" query:"pageSize"` // 每页数量
}

func (o *OrderParam) GetOrderCondition() string {
	return o.Column + " " + o.Order
}

// CalculatePagination 计算并返回分页的 offset 和 limit.
func CalculatePagination(pageParam PageParam) (int, int) {
	var offset, limit int
	if pageParam.PageSize > 0 {
		limit = pageParam.PageSize
		if pageParam.PageNum > 0 {
			offset = (pageParam.PageNum - 1) * pageParam.PageSize
		}
	}
	return offset, limit
}

// https://github.com/ulid/spec
// uuid sortable by time.
// nolint:gosec
// 创建一个新的ULID
func NewUUID() string {
	now := time.Now()
	return ulid.MustNew(ulid.Timestamp(now), ulid.Monotonic(rand.New(rand.NewSource(now.UnixNano())), 0)).String()
}

// RecordExists 通用存在性检查函数.
func RecordExists(query *gorm.DB) (bool, error) {
	// 使用查询创建适当的 DB 对象并检查记录是否存在
	var count int64
	err := query.Limit(1).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("counting records: %w", err)
	}
	return count > 0, nil
}
