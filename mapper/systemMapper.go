package mapper

import "github.com/jimu-server/model"

type SystemMapper struct {
	// 查询所以用户信息
	AllUserList func(any) ([]*model.User, int64, error)
}
