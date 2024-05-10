package mapper

import (
	"database/sql"
	"github.com/jimu-server/model"
)

type AccountMapper struct {
	Register   func(model.User, *sql.Tx) error
	IsRegister func(model.User) (bool, error)

	SelectAccount  func(any) (model.User, error)
	SelectUserById func(any) (model.User, error)

	UpdateUserName          func(any, *sql.Tx) error
	UpdateUserGender        func(any, *sql.Tx) error
	UpdateUserAge           func(any, *sql.Tx) error
	UpdateUserAvatar        func(any) error
	UpdateUserPassword      func(any) error
	RestUserPasswordByPhone func(any) error
	RestUserPasswordByEmail func(any) error
	UpdateUserPhone         func(any) error
	UpdateUserEmail         func(any) error

	UpdateUserOrgRole func(any, *sql.Tx) error
	UpdateUserOrg     func(any, *sql.Tx) error

	CheckUserPhone func(any) (bool, error)
	CheckUserEmail func(any) (bool, error)

	SettingsList  func(any) ([]*model.AppSetting, error)
	GetSettingIds func(any) ([]string, error)
}
