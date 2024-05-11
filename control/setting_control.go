package control

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jimu-server/common/resp"
	"github.com/jimu-server/middleware/auth"
	"github.com/jimu-server/model"
	"github.com/jimu-server/redis/redisUtil"
	"github.com/jimu-server/setting"
	"github.com/jimu-server/util/treeutils/tree"
	"github.com/jimu-server/web"
)

func GetSettings(c *gin.Context) {
	var err error
	var reqParams *SettingsArgs
	web.BindJSON(c, &reqParams)
	token := c.MustGet(auth.Key).(*auth.Token)
	// 缓存查询
	if data := setting.QueryUserSetting(token.Id); data != nil {
		c.JSON(200, resp.Success(data, resp.Msg("获取成功")))
		return
	}
	param := map[string]any{
		"list":   reqParams.Tools,
		"UserId": token.Id,
	}

	// 再次查询出用户的配置
	var set []*model.AppSetting
	if set, err = AccountMapper.SettingsList(param); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("获取失败")))
		return
	}
	// 添加系统默认的 用户个人设置项
	var userSet *model.AppSetting
	if userSet, err = AccountMapper.GetUserInfoSetting(); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("获取失败")))
		return
	}
	result := make([]*model.AppSetting, 0)
	result = append(result, userSet)
	result = append(result, set...)
	buildTree := tree.BuildTree("", result)
	// 放入缓存
	if err = setting.UpdateUserSetting(token.Id, buildTree); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("获取失败")))
		return
	}
	c.JSON(200, resp.Success(buildTree, resp.Msg("获取成功")))
}

func CreateSetting(id string) {

}

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
	if err = AccountMapper.UpdateSetting(param); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("更新失败")))
		return
	}
	// 删除缓存
	if err = redisUtil.Del(fmt.Sprintf("%s:%s", setting.USER_SETTING, token.Id)); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("更新失败")))
		return
	}
}
