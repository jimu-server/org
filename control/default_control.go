package control

import (
	"github.com/gin-gonic/gin"
	"github.com/jimu-server/common/resp"
	"github.com/jimu-server/db"
	"github.com/jimu-server/middleware/auth"
	"github.com/jimu-server/model"
	"github.com/jimu-server/org/dao"

	"net/http"
)

func GetUserDefaultInfo(c *gin.Context) {
	token := c.MustGet(auth.Key).(*auth.Token)
	org := &model.Org{}
	role := &model.Role{}
	var err error
	params := map[string]any{
		"UserId": token.Id,
	}
	if org, err = dao.DefaultInfoMapper.SelectUserDefaultOrg(params); err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err, resp.Msg("获取默认组织失败")))
		return
	}
	params["OrgId"] = org.Id
	if role, err = dao.DefaultInfoMapper.SelectUserDefaultRole(params); err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err, resp.Msg("获取默认组织角色失败")))
		return
	}
	c.JSON(http.StatusOK, resp.Success(map[string]any{
		"org":  org,
		"role": role,
	}, resp.Msg("获取默认信息成功")))
}

func GetOrgDefaultRole(c *gin.Context) {
	var err error
	role := &model.Role{}
	token := c.MustGet(auth.Key).(*auth.Token)
	orgId := c.Query("orgId")
	params := map[string]any{
		"UserId": token.Id,
		"OrgId":  orgId,
	}
	if role, err = dao.DefaultInfoMapper.SelectUserDefaultRole(params); err != nil {
		c.JSON(http.StatusInternalServerError, resp.Error(err, resp.Msg("获取默认组织角色失败")))
		return
	}
	c.JSON(http.StatusOK, resp.Success(role, resp.Msg("获取默认组织角色成功")))
}

func UserInfo(c *gin.Context) {
	token := c.MustGet(auth.Key).(*auth.Token)
	var user *model.User
	db.Gorm.Find(&user, "id =?", token.Id).First(&user)
	c.JSON(http.StatusOK, resp.Success(user, resp.Msg("获取用户信息成功")))
}
