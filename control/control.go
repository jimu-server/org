package control

import (
	"github.com/jimu-server/logger"
	"github.com/jimu-server/org/mapper"
)

const (
	// ROOT_ORG_ID 跟组织ID 数据库初始化脚本中已经初始化
	ROOT_ORG_ID = "1"
	// ROOT_ORG_DEFAULT_ROLE 跟组织默认角色ID 数据库初始化脚本中已经初始化
	ROOT_ORG_DEFAULT_ROLE = "3"

	// GPT_TOOL_ID  数据库初始化脚本中已经初始化 app_tool 表中
	GPT_TOOL_ID = "2"
)

var (
	AccountMapper     = &mapper.AccountMapper{}
	OrgMapper         = &mapper.OrgMapper{}
	RoleMapper        = &mapper.RoleMapper{}
	ToolMapper        = &mapper.ToolMapper{}
	FunMapper         = &mapper.FunMapper{}
	AuthMapper        = &mapper.AuthMapper{}
	DefaultInfoMapper = &mapper.DefaultInfoMapper{}
	MenuMapper        = &mapper.MenuMapper{}
	SystemMapper      = &mapper.SystemMapper{}

	logs = logger.Logger
)
