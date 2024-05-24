package control

import (
	"github.com/gin-gonic/gin"
	"github.com/jimu-server/common/resp"
	"github.com/jimu-server/model"
	"github.com/jimu-server/org/dao"
	"net/http"
)

// AllMenu godoc
// @Summary      获取菜单
// @Description  获取前端路由菜单树
// @Tags         web
// @Accept       json
// @Produce      json
// @Success      200  {object}  resp.Response{data=[]model.Router}
// @Failure      500  {object}  resp.Response
// @Router       /menu [get]
func AllMenu(c *gin.Context) {
	var err error
	var menus []*model.Router
	if menus, err = dao.MenuMapper.SelectAllMenu(); err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err, resp.Msg("获取菜单失败")))
		return
	}
	c.JSON(http.StatusOK, resp.Success(menus, resp.Msg("获取菜单成功")))
}
