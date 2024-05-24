package service

import "github.com/jimu-server/org/dao"

// GetOrgUserAllRole
// 获取用户指定组织下的所有角色信息
func GetOrgUserAllRole(userId string, orgId string) ([]string, error) {
	params := map[string]any{
		"userId": userId,
		"orgId":  orgId,
	}
	var err error
	var roles []string
	if roles, err = dao.AuthMapper.QueryOrgUserRoleIdList(params); err != nil {
		return nil, err
	}
	return roles, nil
}

// GetOrgUserRoleAllTool
// 获取组织用户的所有工具栏id
func GetOrgUserRoleAllTool(userId string, orgId string) ([]string, error) {
	var tools, roles []string
	var err error
	if roles, err = GetOrgUserAllRole(userId, orgId); roles == nil {
		return nil, err
	}
	params := map[string]any{
		"userId": userId,
		"orgId":  orgId,
		"roles":  roles,
	}
	if tools, err = dao.AuthMapper.QueryOrgUserToolIdList(params); err != nil {
		return nil, err
	}
	return tools, nil
}

// GetOrgUserToolAllRout
// 查找用户所有已授权路由
func GetOrgUserToolAllRout(userId string, orgId string) ([]string, error) {
	var menus, roles []string
	var err error
	if roles, err = GetOrgUserAllRole(userId, orgId); roles == nil {
		return nil, err
	}
	params := map[string]any{
		"userId": userId,
		"orgId":  orgId,
		"roles":  roles,
	}
	if menus, err = dao.AuthMapper.QueryOrgUserRouterIdList(params); err != nil {
		return nil, err
	}
	return menus, nil
}
