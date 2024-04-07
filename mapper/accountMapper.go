package mapper

import (
	"database/sql"
	"github.com/jimu-server/model"
)

type AccountMapper struct {
	Register   func(model.User, *sql.Tx) error
	IsRegister func(user model.User) (bool, error)

	SelectAccount func(any) (model.User, error)
}
