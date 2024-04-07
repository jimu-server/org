package mapper

import (
	"github.com/jimu-server/model"
)

type FunMapper struct {
	// 分页查询 子组织
	GetFun func(any) ([]*model.FunRouter, int64, error)
	// 创建工具
	CreateFun func(any) error
	// 删除工具
	DeleteFun func(any) error
	// 更新工具
	UpdateFun func(any) error
}
