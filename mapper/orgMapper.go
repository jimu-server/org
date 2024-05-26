package mapper

import (
	"database/sql"
	"github.com/jimu-server/model"
)

type OrgMapper struct {
	// 查询所有组织
	AllOrg func(any) ([]*model.Org, error)
	// 分页查询 子组织
	GetOrgChild       func(any) ([]*model.Org, int64, error)
	GetOrgUserList    func(any) ([]*model.User, int64, error)
	GetOrgAllUserList func(any) ([]*model.User, int64, error)
	GetOrgRoleList    func(any) ([]*model.Role, int64, error)

	// 创建组织
	CreateOrg func(any) error
	// 删除组织
	DeleteOrg func(any) error

	// 组织添加用户
	OrgAddUser func(any, *sql.Tx) error

	// 查询组织是否有子组织
	IsParentOrg func(any) ([]string, error)

	ExistUser func(any) (string, error)

	UpdateOrg func(any) error

	GetDictionary func() ([]*model.AppDictionary, error)
}
