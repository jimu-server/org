package mapper

import (
	"database/sql"
	"github.com/jimu-server/model"
)

type AccountMapper struct {
	Register   func(model.User, *sql.Tx) error
	IsRegister func(user model.User) (bool, error)

	SelectAccount func(any) (model.User, error)

	UpdateUserName   func(any, *sql.Tx) error
	UpdateUserGender func(any, *sql.Tx) error
	UpdateUserAge    func(any, *sql.Tx) error
	UpdateUserAvatar func(any) error

	UpdateUserOrgRole func(any, *sql.Tx) error
}
