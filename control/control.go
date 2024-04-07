package control

import (
	"github.com/jimu-server/logger"
	"github.com/jimu-server/org/mapper"
)

var AccountMapper = &mapper.AccountMapper{}
var OrgMapper = &mapper.OrgMapper{}
var RoleMapper = &mapper.RoleMapper{}
var ToolMapper = &mapper.ToolMapper{}
var FunMapper = &mapper.FunMapper{}
var AuthMapper = &mapper.AuthMapper{}
var DefaultInfoMapper = &mapper.DefaultInfoMapper{}
var MenuMapper = &mapper.MenuMapper{}
var SystemMapper = &mapper.SystemMapper{}

var logs = logger.Logger
