package control

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jimu-server/common/resp"
	"github.com/jimu-server/model"
	"github.com/jimu-server/util/pageutils"
	"github.com/jimu-server/util/uuidutils/uuid"
	"github.com/jimu-server/web"
)

func CreateFun(c *gin.Context) {
	var args *model.FunRouter
	var err error
	web.BindJSON(c, &args)
	args.Id = uuid.String()
	if err = FunMapper.CreateFun(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("创建失败")))
		return
	}
	c.JSON(200, resp.Success(args, resp.Msg("创建成功")))
}

func DeleteFun(c *gin.Context) {
	var args *DelArgs
	var err error
	web.ShouldJSON(c, &args)
	if args.List == nil || len(args.List) == 0 {
		c.JSON(500, resp.Error(errors.New("id错误"), resp.Msg("删除失败")))
		return
	}

	// todo 判断当前的 功能的授权情况以后在进行删除

	if err = FunMapper.DeleteFun(map[string]any{
		"list": args.List,
	}); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("删除失败")))
		return
	}
	//删除角色
	c.JSON(200, resp.Success(nil, resp.Msg("删除成功")))
}

func GetFun(c *gin.Context) {
	var err error
	var orgs []*model.FunRouter
	var args ListArgs
	web.ShouldJSON(c, &args)
	limit, offset := 0, 0
	if limit, offset, err = pageutils.PageNumber(args.Page, args.PageSize); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("分页参数错误")))
		return
	}
	args.Start, args.End = offset, limit
	var count int64 = 0
	if orgs, count, err = FunMapper.GetFun(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	page := resp.NewPage(count, orgs)
	c.JSON(200, resp.Success(page, resp.Msg("查询成功")))
}
