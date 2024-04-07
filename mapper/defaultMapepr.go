package mapper

import (
	"database/sql"
	"github.com/jimu-server/model"
)

type DefaultInfoMapper struct {
	// 查询用户默认组织
	SelectUserDefaultOrg func(any) (*model.Org, error)

	// 查询用户默认角色
	SelectUserDefaultRole func(any) (*model.Role, error)

	// 设置用户默认组织
	SetUserDefaultOrg func(any, *sql.Tx) error

	// 设置用户默认角色
	SetUserDefaultRole func(any, *sql.Tx) error

	SelectUserInfo func(any) (*model.User, error)
}
