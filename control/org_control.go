package control

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jimu-server/common/resp"
	"github.com/jimu-server/db"
	"github.com/jimu-server/model"
	"github.com/jimu-server/org/dao"
	"github.com/jimu-server/util/pageutils"
	"github.com/jimu-server/util/treeutils/tree"
	"github.com/jimu-server/util/uuidutils/uuid"
	"github.com/jimu-server/web"
	"net/http"
)

func CreateOrg(c *gin.Context) {
	var args *model.Org
	var err error
	if err = c.BindJSON(&args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("请求参数解析失败")))
		return
	}
	args.Id = uuid.String()
	if err = dao.OrgMapper.CreateOrg(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("创建失败")))
		return
	}
	c.JSON(200, resp.Success(args, resp.Msg("创建成功")))
}

func DeleteOrg(c *gin.Context) {
	var args *model.Org
	var err error
	if err = c.BindJSON(&args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("请求参数解析失败")))
		return
	}
	if args.Id == "" {
		c.JSON(500, resp.Error(errors.New("组织id错误"), resp.Msg("删除失败")))
		return
	}
	// 1.查询组织内是否存在用户
	id := ""
	if id, err = dao.OrgMapper.ExistUser(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("删除失败")))
		return
	}
	if id != "" {
		c.JSON(500, resp.Error(errors.New("组织存在未移除用户"), resp.Msg("删除失败")))
		return
	}
	// 2.检查是否有子组织
	var ids []string
	if ids, err = dao.OrgMapper.IsParentOrg(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("删除失败")))
		return
	}
	if ids != nil && len(ids) > 0 {
		c.JSON(500, resp.Error(errors.New("存在子组织"), resp.Msg("删除失败")))
		return
	}
	// 3.删除组织
	if err = dao.OrgMapper.DeleteOrg(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("组织删除失败")))
		return
	}
	//删除组织
	c.JSON(200, resp.Success(nil, resp.Msg("组织删除成功")))
	return
}

func GetOrg(c *gin.Context) {
	var err error
	var orgs []*model.Org
	var args ListArgs
	if err = c.ShouldBind(&args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("请求参数解析失败")))
		return
	}
	limit, offset := 0, 0
	if limit, offset, err = pageutils.PageNumber(args.Page, args.PageSize); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("分页参数错误")))
		return
	}
	args.Start, args.End = offset, limit
	var count int64 = 0
	if orgs, count, err = dao.OrgMapper.GetOrgChild(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	page := resp.NewPage(count, orgs)
	c.JSON(200, resp.Success(page, resp.Msg("查询成功")))
	return
}

func GetOrgUserList(c *gin.Context) {
	var err error
	var users []*model.User
	var args OrgUserListArgs
	web.ShouldJSON(c, &args)
	limit, offset := 0, 0
	if limit, offset, err = pageutils.PageNumber(args.Page, args.PageSize); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("分页参数错误")))
		return
	}
	args.Start, args.End = offset, limit
	var count int64 = 0
	if users, count, err = dao.OrgMapper.GetOrgUserList(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	page := resp.NewPage(count, users)
	c.JSON(200, resp.Success(page, resp.Msg("查询成功")))
	return
}

func GetOrgRoleList(c *gin.Context) {
	var err error
	var roles []*model.Role
	var args OrgRoleListArgs
	web.ShouldJSON(c, &args)
	limit, offset := 0, 0
	if limit, offset, err = pageutils.PageNumber(args.Page, args.PageSize); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("分页参数错误")))
		return
	}
	args.Start, args.End = offset, limit
	var count int64 = 0
	if roles, count, err = dao.OrgMapper.GetOrgRoleList(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	page := resp.NewPage(count, roles)
	c.JSON(200, resp.Success(page, resp.Msg("查询成功")))
}

func UpdateOrgInfo(c *gin.Context) {
	var err error
	var args *UpdateOrg
	if err = c.BindJSON(&args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("请求参数解析失败")))
		return
	}
	if err = dao.OrgMapper.UpdateOrg(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("更新失败")))
		return
	}
	c.JSON(200, resp.Success(nil, resp.Msg("更新成功")))
}

func Dictionary(c *gin.Context) {
	var err error
	var dict []*model.AppDictionary
	if dict, err = dao.OrgMapper.GetDictionary(); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	dicts := make(map[string][]any)
	for _, v := range dict {
		dicts[v.Type] = append(dicts[v.Type], v)
	}
	c.JSON(200, resp.Success(dicts, resp.Msg("查询成功")))
}

func GetOrgUserAuthTool(c *gin.Context) {
	var err error
	roleId := c.Query("roleId")
	orgId := c.Query("orgId")
	userId := c.Query("userId")
	params := map[string]any{
		"RoleId": roleId,
		"UserId": userId,
		"OrgId":  orgId,
	}
	tools := []*model.Tool{}
	if tools, err = dao.AuthMapper.SelectOrgUserAuthTool(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	c.JSON(200, resp.Success(tools))
}

/*
	func GetOrgRoleAuthToolRouter(c *gin.Context) {
		var err error
		roleId := c.Query("roleId")
		orgId := c.Query("orgId")
		userId := c.Query("userId")
		toolId := c.Query("toolId")
		params := map[string]any{
			"RoleId": roleId,
			"UserId": userId,
			"OrgId":  orgId,
			"ToolId": toolId,
		}
		routers := []*model.Router{}
		if routers, err = AuthMapper.SelectOrgRoleAuthToolRouter(params); err != nil {
			logs.Error(err.Error())
			c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
			return
		}
		routerTree := tree.BuildTree("", routers)
		c.JSON(200, resp.Success(routerTree))
	}
*/
func OrgUserRoleAuth(c *gin.Context) {
	var err error
	var args *AuthArgs
	params := map[string]any{}
	web.BindJSON(c, &args)
	begin, err := db.DB.Begin()
	if err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("开启事务失败")))
		return
	}

	// 1.给用户添加角色授权
	if len(args.Auths) != 0 {
		var list []model.AuthUserRole
		for _, v := range args.Auths {
			list = append(list, model.AuthUserRole{Id: uuid.String(), RoleId: v, UserId: args.UserId, OrgId: args.OrgId, FirstChoice: false})
		}
		params["list"] = list
		if err = dao.AuthMapper.AddOrgUserRoleAuth(params, begin); err != nil {
			c.JSON(500, resp.Error(err, resp.Msg("添加失败")))
			return
		}
	}

	// 2.给用户授权对应角色的所有权限
	var authTool []model.AuthTool
	var toolIds []string
	var authToolRouter []model.AuthRouter
	var toolRouterIds []string
	queryParams := map[string]any{}
	queryParams["OrgId"] = args.OrgId
	for _, v := range args.Auths {
		queryParams["RoleId"] = v
		// 查询对应角色的所有工具栏权限
		if toolIds, err = dao.AuthMapper.SelectOrgRoleToolAuth(queryParams); err != nil {
			begin.Rollback()
			c.JSON(500, resp.Error(err, resp.Msg("工具栏权限授权失败")))
			return
		}
		for _, toolId := range toolIds {
			// 封装工具栏授权参数
			authTool = append(authTool, model.AuthTool{Id: uuid.String(), RoleId: v, UserId: args.UserId, OrgId: args.OrgId, ToolId: toolId})

			queryParams["ToolId"] = toolId
			// 查询对应角色对应工具栏的所有路由权限
			if toolRouterIds, err = dao.AuthMapper.SelectOrgRoleRouterAuth(queryParams); err != nil {
				begin.Rollback()
				c.JSON(500, resp.Error(err, resp.Msg("工具栏权限授权失败")))
				return
			}
			// 封装工具栏路由授权参数
			for _, toolRouterId := range toolRouterIds {
				authToolRouter = append(authToolRouter, model.AuthRouter{Id: uuid.String(), OrgId: args.OrgId, UserId: args.UserId, RoleId: v, ToolId: toolId, RouterId: toolRouterId})
			}
		}
	}

	// 2.1 开始授权工具
	if len(authTool) != 0 {
		params["list"] = authTool
		if err = dao.AuthMapper.AddOrgUserRoleToolAuth(params, begin); err != nil {
			begin.Rollback()
			c.JSON(500, resp.Error(err, resp.Msg("工具栏权限授权失败")))
			return
		}
	}

	// 2.2 开始工具栏路由授权
	if len(authToolRouter) != 0 {
		params["list"] = authToolRouter
		if err = dao.AuthMapper.AddOrgUserRoleToolRouterAuth(params, begin); err != nil {
			begin.Rollback()
			c.JSON(500, resp.Error(err, resp.Msg("工具栏路由权限授权失败")))
			return
		}
	}

	// 2.3 授权功能路由

	// 3.给角色取消授权
	if len(args.UnAuth) != 0 {
		// 2.给角色取消授权
		if err = dao.AuthMapper.DeleteOrgUserRoleAuth(map[string]any{
			"OrgId":  args.OrgId,
			"UserId": args.UserId,
			"list":   args.UnAuth,
		}, begin); err != nil {
			begin.Rollback()
			c.JSON(500, resp.Error(err, resp.Msg("删除失败")))
			return
		}
	}
	// 4. 删除组织用户取消角色对应的所有权限
	var unAuthTool map[string]any
	var unToolIds []string
	var unAuthToolRouter []map[string]any
	for _, roleId := range args.UnAuth {
		// 4.1 删除组织用户取消角色对应的 工具栏授权
		if unToolIds, err = dao.AuthMapper.SelectOrgRoleToolAuth(map[string]any{
			"RoleId": roleId,
			"UserId": args.UserId,
			"OrgId":  args.OrgId,
		}); err != nil {
			begin.Rollback()
			logs.Error(err.Error())
			c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
			return
		}

		// 封装待删除授权的工具sql参数 删除 toolIds 工具权限
		unAuthTool = map[string]any{
			"RoleId": roleId,
			"UserId": args.UserId,
			"OrgId":  args.OrgId,
			"list":   unToolIds,
		}
		// 4.2 删除组织用户取消工具栏对应的功能路由授权
		for _, toolId := range unToolIds {
			var routerIds []string
			// 查询组织用户角色对应的工具栏的路由权限
			if routerIds, err = dao.AuthMapper.SelectOrgRoleRouterAuth(map[string]any{
				"OrgId":  args.OrgId,
				"UserId": args.UserId,
				"RoleId": roleId,
				"ToolId": toolId,
			}); err != nil {
				begin.Rollback()
				logs.Error(err.Error())
				c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
				return
			}
			if len(routerIds) != 0 {
				unAuthToolRouter = append(unAuthToolRouter, map[string]any{
					"RoleId": roleId,
					"UserId": args.UserId,
					"OrgId":  args.OrgId,
					"ToolId": toolId,
					"list":   routerIds,
				})
			}
		}
	}

	if len(unToolIds) != 0 {
		if err = dao.AuthMapper.DelOrgUserRoleToolAuth(unAuthTool, begin); err != nil {
			begin.Rollback()
			c.JSON(500, resp.Error(err, resp.Msg("删除失败")))
			return
		}
	}

	if len(unAuthToolRouter) != 0 {
		for _, unauth := range unAuthToolRouter {
			if err = dao.AuthMapper.DelOrgUserRoleToolRouterAuth(unauth, begin); err != nil {
				begin.Rollback()
				c.JSON(500, resp.Error(err, resp.Msg("删除失败")))
				return
			}
		}
	}
	begin.Commit()
	c.JSON(200, resp.Success(nil))
}

func OrgUserRoleToolAuth(c *gin.Context) {
	var err error
	var args *AuthArgs
	web.BindJSON(c, &args)
	begin, err := db.DB.Begin()
	if err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("开启事务失败")))
		return
	}
	// 检查是否有需要授权的工具
	if len(args.Auths) != 0 {
		var list []model.AuthTool
		for ids := range args.Auths {
			list = append(list, model.AuthTool{Id: uuid.String(), UserId: args.UserId, OrgId: args.OrgId, ToolId: args.Auths[ids], RoleId: args.RoleId})
		}
		params := map[string]any{
			"list": list,
		}
		if err = dao.AuthMapper.AddOrgUserRoleToolAuth(params, begin); err != nil {
			begin.Rollback()
			c.JSON(500, resp.Error(err, resp.Msg("添加失败")))
			return
		}
	}
	// 检查是否有需要取消授权的工具
	if len(args.UnAuth) != 0 {
		params := map[string]any{
			"OrgId":  args.OrgId,
			"UserId": args.UserId,
			"RoleId": args.RoleId,
			"list":   args.UnAuth,
		}
		if err = dao.AuthMapper.DelOrgUserRoleToolAuth(params, begin); err != nil {
			begin.Rollback()
			c.JSON(500, resp.Error(err, resp.Msg("删除失败")))
			return
		}
	}
	begin.Commit()
	c.JSON(http.StatusOK, resp.Success(nil))
}

func OrgUserRoleToolRoleAuth(c *gin.Context) {
	var err error
	var args *AuthArgs
	web.BindJSON(c, &args)
	begin, err := db.DB.Begin()
	if err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("开启事务失败")))
		return
	}
	if len(args.Auths) != 0 {
		var list []model.AuthRouter
		for ids := range args.Auths {
			list = append(list, model.AuthRouter{Id: uuid.String(), UserId: args.UserId, OrgId: args.OrgId, ToolId: args.ToolId, RoleId: args.RoleId, RouterId: args.Auths[ids]})
		}
		params := map[string]any{
			"list": list,
		}
		if err = dao.AuthMapper.AddOrgUserRoleToolRouterAuth(params, begin); err != nil {
			begin.Rollback()
			c.JSON(500, resp.Error(err, resp.Msg("添加失败")))
			return
		}
	}

	if len(args.UnAuth) != 0 {
		params := map[string]any{
			"OrgId":  args.OrgId,
			"UserId": args.UserId,
			"RoleId": args.RoleId,
			"ToolId": args.ToolId,
			"list":   args.UnAuth,
		}
		if err = dao.AuthMapper.DelOrgUserRoleToolRouterAuth(params, begin); err != nil {
			begin.Rollback()
			c.JSON(500, resp.Error(err, resp.Msg("删除失败")))
			return
		}
	}
	begin.Commit()
	c.JSON(http.StatusOK, resp.Success(nil))
}

func GetOrgUserAuthToolRouter(c *gin.Context) {
	var err error
	roleId := c.Query("roleId")
	orgId := c.Query("orgId")
	userId := c.Query("userId")
	toolId := c.Query("toolId")
	params := map[string]any{
		"RoleId": roleId,
		"UserId": userId,
		"OrgId":  orgId,
		"ToolId": toolId,
	}
	routers := []*model.Router{}
	if routers, err = dao.AuthMapper.SelectOrgUserAuthToolRouter(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	routerTree := tree.BuildTree("", routers)
	c.JSON(200, resp.Success(routerTree))
}
