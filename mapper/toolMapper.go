package mapper

import (
	"database/sql"
	"github.com/jimu-server/model"
)

type ToolMapper struct {
	// 分页查询 工具
	GetTool func(any) ([]*model.Tool, int64, error)
	// 分页获取顶层路由
	GetToolRouter func(any) ([]*model.Router, int64, error)
	// 获取子路由
	GetToolRouterChild func(any) ([]*model.Router, error)
	// 创建工具
	CreateTool func(any) error
	// 检查工具路径组件是否重复
	CheckTool func(any) (bool, error)

	// 删除工具
	DeleteTool func(any) error
	// 更新工具
	UpdateTool func(any) error

	ToolStatus func(any, *sql.Tx) error
}
