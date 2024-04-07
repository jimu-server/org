package mapper

import (
	"github.com/jimu-server/model"
)

type MenuMapper struct {
	SelectAllMenu func() ([]*model.Router, error)
}
