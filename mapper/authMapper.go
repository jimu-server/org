package mapper

import (
	"database/sql"
	"github.com/jimu-server/model"
)

type AuthMapper struct {
	SelectAuthUserMenu func(any) ([]*model.Router, error)

	SelectAuthUserTool func(any) ([]*model.Tool, error)

	// 获取当前用户的已授权工具栏(不区分工具栏位置)
	SelectOrgUserAuthTool func(any) ([]*model.Tool, error)

	SelectOrgUserAuthToolRouter func(any) ([]*model.Router, error)

	SelectAuthUserToolMenu      func(any) ([]*model.Router, error)
	SelectAuthUserToolMenuChild func(any) ([]*model.Router, error)

	SelectAuthAllUserRouterPath func(any) ([]string, error)

	SelectAuthAllUserToolRouterPath func(any) ([]string, error)

	SelectUserOrgList func(any) ([]*model.Org, error)
	SelectAllOrg      func() ([]*model.Org, error)

	SelectUserOrgRoleList func(any) ([]*model.Role, error)

	// 对组织的用户户进行角色授权
	AddOrgUserRoleAuth         func(any, *sql.Tx) error
	RegisterAddOrgUserRoleAuth func(any, *sql.Tx) error
	DeleteOrgUserRoleAuth      func(any, *sql.Tx) error

	// 查询组织对应角色的工具栏授权
	SelectOrgRoleToolAuth func(any) ([]string, error)

	// 查询组织对应角色对应工具栏的路由授权
	SelectOrgRoleRouterAuth func(any) ([]string, error)

	// 添加组织用户对应角色的工具栏权限
	AddOrgUserRoleToolAuth func(any, *sql.Tx) error
	DelOrgUserRoleToolAuth func(any, *sql.Tx) error

	// 添加组织用户对应角色对应工具栏路由权限
	AddOrgUserRoleToolRouterAuth func(any, *sql.Tx) error
	DelOrgUserRoleToolRouterAuth func(any, *sql.Tx) error

	OrgRoleToolList       func(any) ([]*model.Tool, error)
	OrgRoleToolRouterList func(any) ([]*model.Router, error)

	OrgRoleToolAuth       func(any, *sql.Tx) error
	OrgRoleToolRouterAuth func(any, *sql.Tx) error

	OrgRoleToolUnAuth       func(any, *sql.Tx) error
	OrgRoleToolRouterUnAuth func(any, *sql.Tx) error
}
