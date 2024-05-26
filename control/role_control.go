package control

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jimu-server/common/resp"
	"github.com/jimu-server/db"
	"github.com/jimu-server/middleware/auth"
	"github.com/jimu-server/model"
	"github.com/jimu-server/org/dao"
	"github.com/jimu-server/util/pageutils"
	"github.com/jimu-server/util/treeutils/tree"
	"github.com/jimu-server/util/uuidutils/uuid"
	"github.com/jimu-server/web"
)

func CreateRole(c *gin.Context) {
	var args *CreateOrgRole
	var err error
	token := c.MustGet(auth.Key).(*auth.Token)
	web.BindJSON(c, &args)
	var begin *sql.Tx
	if begin, err = db.DB.Begin(); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("开启事务失败")))
		return
	}
	role := model.Role{
		Id:      uuid.String(),
		Name:    args.Name,
		RoleKey: args.RoleKey,
	}
	// 创建角色
	if err = dao.RoleMapper.CreateRole(role, begin); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("创建失败")))
		return
	}
	// 创建关联关系
	args.Id = uuid.String()
	args.RoleId = role.Id
	if err = dao.RoleMapper.CreateOrgRole(args, begin); err != nil {
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("创建失败")))
		return
	}
	// 给创建者授权该角色
	auth := model.AuthUserRole{
		Id:          uuid.String(),
		UserId:      token.Id,
		RoleId:      role.Id,
		OrgId:       args.OrgId,
		FirstChoice: false,
	}
	params := map[string]interface{}{
		"list": []model.AuthUserRole{auth},
	}
	if err = dao.AuthMapper.AddOrgUserRoleAuth(params, begin); err != nil {
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("授权失败")))
		return
	}
	begin.Commit()
	c.JSON(200, resp.Success(args, resp.Msg("创建成功")))
}

func DeleteRole(c *gin.Context) {
	var args *model.Role
	var err error
	if err = c.BindJSON(&args); err != nil {
		panic(web.ArgsErr())
	}
	if args.Id == "" {
		c.JSON(500, resp.Error(errors.New("角色id错误"), resp.Msg("删除失败")))
		return
	}

	//todo 1.检查角色是否处于在使用状态
	//id := ""
	//if id, err = RoleMapper.ExistUser(args); err != nil {
	//	c.JSON(500, resp.Error(err, resp.Msg("删除失败")))
	//	return
	//}
	//if id != "" {
	//	c.JSON(500, resp.Error(errors.New("角色存在未移除用户"), resp.Msg("删除失败")))
	//	return
	//}

	// 3.删除角色
	if err = dao.RoleMapper.DeleteRole(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("角色删除失败")))
		return
	}
	//删除角色
	c.JSON(200, resp.Success(nil, resp.Msg("角色删除成功")))
	return
}

func GetRole(c *gin.Context) {
	var err error
	var orgs []*model.Role
	var args ListArgs
	if err = c.ShouldBind(&args); err != nil {
		panic(web.ArgsErr())
	}
	limit, offset := 0, 0
	if limit, offset, err = pageutils.PageNumber(args.Page, args.PageSize); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("分页参数错误")))
		return
	}
	args.Start, args.End = offset, limit
	var count int64 = 0
	if orgs, count, err = dao.RoleMapper.GetRole(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	page := resp.NewPage(count, orgs)
	c.JSON(200, resp.Success(page, resp.Msg("查询成功")))
	return
}

func UpdateRoleInfo(c *gin.Context) {
	var err error
	var args *UpdateRole
	if err = c.BindJSON(&args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("请求参数解析失败")))
		return
	}
	if err = dao.RoleMapper.UpdateRole(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("更新失败")))
		return
	}
	c.JSON(200, resp.Success(nil, resp.Msg("更新成功")))
}

func OrgRoleToolList(c *gin.Context) {
	var err error
	var tools []*model.Tool
	var args *RoleAuthQuery
	web.ShouldJSON(c, &args)
	if tools, err = dao.AuthMapper.OrgRoleToolList(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	c.JSON(200, resp.Success(tools, resp.Msg("查询成功")))
}

// OrgRoleToolRouterList
// @Summary 	获取角色已授权的工具栏对应的路由列表
// @Description 获取角色已授权的工具栏对应的路由列表
// @Tags 		管理系统
// @Accept 		json
// @Produces 	json
// @Param 		args body RoleAuthQuery true "请求体"
// @Router 		/api/role/tool/router/tree [get]
func OrgRoleToolRouterList(c *gin.Context) {
	var err error
	var routers []*model.Router
	var args *RoleAuthQuery
	web.ShouldJSON(c, &args)
	if routers, err = dao.AuthMapper.OrgRoleToolRouterList(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	buildTree := tree.BuildTree("", routers)
	c.JSON(200, resp.Success(buildTree, resp.Msg("查询成功")))
}

func OrgRoleToolAuth(c *gin.Context) {
	var err error
	var args *RoleAuthArgs
	web.BindJSON(c, &args)
	var begin *sql.Tx
	if begin, err = db.DB.Begin(); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("开启事务失败")))
		return
	}
	var list []*model.AuthRoleTool
	for auth := range args.Auths {
		list = append(list, &model.AuthRoleTool{
			Id:     uuid.String(),
			OrgId:  args.OrgId,
			RoleId: args.RoleId,
			ToolId: args.Auths[auth],
		})
	}
	params := map[string]any{
		"list": list,
	}
	if len(list) != 0 {
		if err = dao.AuthMapper.OrgRoleToolAuth(params, begin); err != nil {
			c.JSON(500, resp.Error(err, resp.Msg("授权失败")))
			return
		}
	}

	if len(args.UnAuth) != 0 {
		params["list"] = args.UnAuth
		params["OrgId"] = args.OrgId
		params["RoleId"] = args.RoleId
		if err = dao.AuthMapper.OrgRoleToolUnAuth(params, begin); err != nil {
			begin.Rollback()
			c.JSON(500, resp.Error(err, resp.Msg("取消授权失败")))
			return
		}
	}
	begin.Commit()
	c.JSON(200, resp.Success(nil, resp.Msg("授权成功")))
}

func OrgRoleToolRouterAuth(c *gin.Context) {
	var err error
	var args *RoleAuthArgs
	web.BindJSON(c, &args)
	var begin *sql.Tx
	if begin, err = db.DB.Begin(); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("开启事务失败")))
		return
	}
	var list []*model.AuthRoleRouter
	for i := range args.Auths {
		list = append(list, &model.AuthRoleRouter{
			Id:       uuid.String(),
			OrgId:    args.OrgId,
			RoleId:   args.RoleId,
			ToolId:   args.ToolId,
			RouterId: args.Auths[i],
		})
	}
	params := map[string]any{
		"list": list,
	}
	if len(list) != 0 {
		if err = dao.AuthMapper.OrgRoleToolRouterAuth(params, begin); err != nil {
			c.JSON(500, resp.Error(err, resp.Msg("授权失败")))
			return
		}
	}
	if len(args.UnAuth) != 0 {
		params["list"] = args.UnAuth
		params["OrgId"] = args.OrgId
		params["RoleId"] = args.RoleId
		params["ToolId"] = args.ToolId
		if err = dao.AuthMapper.OrgRoleToolRouterUnAuth(params, begin); err != nil {
			begin.Rollback()
			c.JSON(500, resp.Error(err, resp.Msg("取消授权失败")))
			return
		}
		// todo 删除当前组织中所有已授权该角色对应的当前取消授权的路由 (root 除外)
	}
	begin.Commit()
	c.JSON(200, resp.Success(nil, resp.Msg("授权成功")))
}
