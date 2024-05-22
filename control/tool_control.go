package control

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jimu-server/common/resp"
	"github.com/jimu-server/model"
	"github.com/jimu-server/org/dao"
	"github.com/jimu-server/util/pageutils"
	"github.com/jimu-server/util/uuidutils/uuid"
	"github.com/jimu-server/web"
)

func CreateTool(c *gin.Context) {
	var args *model.Tool
	var err error
	web.BindJSON(c, &args)
	var flag bool
	if flag, err = dao.ToolMapper.CheckTool(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("创建失败")))
		return
	} else if flag {
		c.JSON(200, resp.Success(args, resp.Msg("创建失败")))
		return
	}
	args.Id = uuid.String()
	if err = dao.ToolMapper.CreateTool(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("创建失败")))
		return
	}
	c.JSON(200, resp.Success(args, resp.Msg("创建成功")))
}

func DeleteTool(c *gin.Context) {
	var args *DelArgs
	var err error
	web.ShouldJSON(c, &args)
	if args.List == nil || len(args.List) == 0 {
		c.JSON(500, resp.Error(errors.New("id错误"), resp.Msg("删除失败")))
		return
	}

	// todo 判断当前的 工具的授权情况以后在进行删除

	if err = dao.ToolMapper.DeleteTool(map[string]any{
		"list": args.List,
	}); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("删除失败")))
		return
	}
	//删除角色
	c.JSON(200, resp.Success(nil, resp.Msg("删除成功")))
}

func GetTool(c *gin.Context) {
	var err error
	var orgs []*model.Tool
	var args ListArgs
	web.ShouldJSON(c, &args)
	limit, offset := 0, 0
	if limit, offset, err = pageutils.PageNumber(args.Page, args.PageSize); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("分页参数错误")))
		return
	}
	args.Start, args.End = offset, limit
	var count int64 = 0
	if orgs, count, err = dao.ToolMapper.GetTool(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	page := resp.NewPage(count, orgs)
	c.JSON(200, resp.Success(page, resp.Msg("查询成功")))
}

func GetToolRouterList(c *gin.Context) {
	var err error
	var routers []*model.Router
	var args ToolRouterArgs
	web.ShouldJSON(c, &args)
	if args.Pid == "" {
		limit, offset := 0, 0
		if limit, offset, err = pageutils.PageNumber(args.Page, args.PageSize); err != nil {
			c.JSON(500, resp.Error(err, resp.Msg("分页参数错误")))
			return
		}
		args.Start, args.End = offset, limit
		var count int64 = 0
		if routers, count, err = dao.ToolMapper.GetToolRouter(args); err != nil {
			c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
			return
		}
		page := resp.NewPage(count, routers)
		c.JSON(200, resp.Success(page, resp.Msg("查询成功")))
		return
	}
	if routers, err = dao.ToolMapper.GetToolRouterChild(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	c.JSON(200, resp.Success(routers, resp.Msg("查询成功")))
}
