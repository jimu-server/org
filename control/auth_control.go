package control

import (
	"github.com/gin-gonic/gin"
	"github.com/jimu-server/common/resp"
	"github.com/jimu-server/middleware/auth"
	"github.com/jimu-server/model"
	"github.com/jimu-server/org/dao"
	"github.com/jimu-server/util/treeutils/tree"
)

func GetAuthMenu(c *gin.Context) {
	var err error
	token := c.MustGet(auth.Key).(*auth.Token)
	roleId := c.Query("roleId")
	orgId := c.Query("orgId")
	params := map[string]any{
		"RoleId": roleId,
		"UserId": token.Id,
		"OrgId":  orgId,
	}
	menus := []*model.Router{}
	if menus, err = dao.AuthMapper.SelectAuthUserMenu(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	list := tree.BuildTree("", menus)
	c.JSON(200, resp.Success(list))
}

func GetAuthTool(c *gin.Context) {
	var err error
	token := c.MustGet(auth.Key).(*auth.Token)
	roleId := c.Query("roleId")
	orgId := c.Query("orgId")
	position := c.Query("position")
	params := map[string]any{
		"RoleId":   roleId,
		"UserId":   token.Id,
		"OrgId":    orgId,
		"Position": position,
	}
	tools := []*model.Tool{}
	if tools, err = dao.AuthMapper.SelectAuthUserTool(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	c.JSON(200, resp.Success(tools))
}

func GetAuthToolMenu(c *gin.Context) {
	var err error
	token := c.MustGet(auth.Key).(*auth.Token)
	roleId := c.Query("roleId")
	orgId := c.Query("orgId")
	toolId := c.Query("toolId")
	params := map[string]any{
		"RoleId": roleId,
		"UserId": token.Id,
		"OrgId":  orgId,
		"ToolId": toolId,
	}
	menus := []*model.Router{}
	if menus, err = dao.AuthMapper.SelectAuthUserToolMenu(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	list := tree.BuildTree("", menus)
	c.JSON(200, resp.Success(list))
}

func GetAuthToolMenuChild(c *gin.Context) {
	var err error
	token := c.MustGet(auth.Key).(*auth.Token)
	roleId := c.Query("roleId")
	orgId := c.Query("orgId")
	toolId := c.Query("toolId")
	params := map[string]any{
		"RoleId": roleId,
		"UserId": token.Id,
		"OrgId":  orgId,
		"ToolId": toolId,
	}
	menus := []*model.Router{}
	if menus, err = dao.AuthMapper.SelectAuthUserToolMenu(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	list := tree.BuildTree("", menus)
	c.JSON(200, resp.Success(list))
}

func UserAuthAllRoute(c *gin.Context) {
	var err error
	token := c.MustGet(auth.Key).(*auth.Token)
	roleId := c.Query("roleId")
	orgId := c.Query("orgId")
	params := map[string]any{
		"RoleId": roleId,
		"UserId": token.Id,
		"OrgId":  orgId,
	}
	all := []string{}
	list := []string{}
	if list, err = dao.AuthMapper.SelectAuthAllUserRouterPath(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
	}
	all = append(all, list...)
	if list, err = dao.AuthMapper.SelectAuthAllUserToolRouterPath(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	all = append(all, list...)
	c.JSON(200, resp.Success(all))
}

func UserJoinOrgList(c *gin.Context) {
	var err error
	token := c.MustGet(auth.Key).(*auth.Token)
	orgId := c.Query("orgId")
	params := map[string]any{
		"UserId": token.Id,
		"OrgId":  orgId,
	}
	orgs := []*model.Org{}
	if orgs, err = dao.AuthMapper.SelectUserOrgList(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	c.JSON(200, resp.Success(orgs))
}

func UserJoinOrgTreeList(c *gin.Context) {
	var err error
	orgId := c.Query("orgId")
	var all []*model.Org
	if all, err = dao.AuthMapper.SelectAllOrg(); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	trees := tree.Tree(orgId, all)
	c.JSON(200, resp.Success(trees))
}

func UserJoinOrgRoleList(c *gin.Context) {
	var err error
	token := c.MustGet(auth.Key).(*auth.Token)
	orgId := c.Query("orgId")
	params := map[string]any{
		"UserId": token.Id,
		"OrgId":  orgId,
	}
	roles := []*model.Role{}
	if roles, err = dao.AuthMapper.SelectUserOrgRoleList(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	c.JSON(200, resp.Success(roles))
}

func GetOrgUserRoleList(c *gin.Context) {
	var err error
	orgId := c.Query("orgId")
	userId := c.Query("userId")
	params := map[string]any{
		"UserId": userId,
		"OrgId":  orgId,
	}
	roles := []*model.Role{}
	if roles, err = dao.AuthMapper.SelectUserOrgRoleList(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	c.JSON(200, resp.Success(roles))
}
