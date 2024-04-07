package mapper

import (
	"database/sql"
	"github.com/jimu-server/model"
)

type RoleMapper struct {
	// 查询所有组织
	AllRole func(any) ([]*model.Role, error)

	// 分页查询 子组织
	GetRole func(any) ([]*model.Role, int64, error)

	// 创建组织
	CreateRole func(any, *sql.Tx) error

	CreateOrgRole func(any, *sql.Tx) error

	// 删除组织
	DeleteRole func(any) error

	UpdateRole func(any) error
}
