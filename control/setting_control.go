package control

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jimu-server/common/resp"
	"github.com/jimu-server/middleware/auth"
	"github.com/jimu-server/org/dao"
	"github.com/jimu-server/redis/redisUtil"
	"github.com/jimu-server/setting"
	"github.com/jimu-server/web"
)

// GetSettings
// @Summary 	获取用户设置
// @Description 获取用户设置
// @Tags 		用户接口
// @Accept 		json
// @Produces 	json
// @Param 		args body SettingsArgs true "请求体"
// @Router 		/api/settings [post]
func GetSettings(c *gin.Context) {
	var err error
	var reqParams *SettingsArgs
	web.BindJSON(c, &reqParams)
	token := c.MustGet(auth.Key).(*auth.Token)
	// 缓存查询
	var data any
	if data, err = setting.QueryUserSetting(token.Id); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("获取失败")))
		return
	}
	c.JSON(200, resp.Success(data, resp.Msg("获取成功")))
}

// UpdateSettings
// @Summary 	更新用户设置
// @Description 更新用户设置
// @Tags 		用户接口
// @Accept 		json
// @Produces 	json
// @Param 		args body SettingsArgs true "请求体"
// @Router 		/api/settings/update [post]
func UpdateSettings(c *gin.Context) {
	var err error
	var reqParams *SettingsArgs
	web.BindJSON(c, &reqParams)
	token := c.MustGet(auth.Key).(*auth.Token)
	param := map[string]any{
		"Id":      reqParams.SettingId,
		"setting": reqParams.Value,
		"UserId":  token.Id,
	}
	// 更新指定配置项配置数据
	if err = dao.AccountMapper.UpdateSetting(param); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("更新失败")))
		return
	}
	// 删除缓存
	if err = redisUtil.Del(fmt.Sprintf("%s:%s", setting.USER_SETTING, token.Id)); err != nil {
		logs.Error(err.Error())
		return
	}
}
