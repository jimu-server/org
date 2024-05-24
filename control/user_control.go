package control

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jimu-server/common/resp"
	"github.com/jimu-server/config"
	"github.com/jimu-server/db"
	"github.com/jimu-server/middleware/auth"
	"github.com/jimu-server/model"
	"github.com/jimu-server/mq/mq_key"
	"github.com/jimu-server/mq/rabbmq"
	"github.com/jimu-server/org/dao"
	"github.com/jimu-server/oss"
	"github.com/jimu-server/redis/cache"
	"github.com/jimu-server/redis/redisUtil"
	"github.com/jimu-server/setting"
	"github.com/jimu-server/util/accountutil"
	"github.com/jimu-server/util/email163"
	"github.com/jimu-server/util/pageutils"
	"github.com/jimu-server/util/uuidutils/uuid"
	"github.com/jimu-server/web"
	amqp "github.com/rabbitmq/amqp091-go"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Register 注册用户
// @Summary      注册用户
// @Description  系统用户注册
// @Tags         用户接口
// @Accept       json
// @Produce      json
// @Param	     args body RegisterArgs true "请求体"
// @Success      200  {object}  resp.Response{code=int,data=model.User}
// @Failure      500  {object}  resp.Response{code=int,data=any}
// @Router       /public/register [post]
func Register(c *gin.Context) {
	var body RegisterArgs
	var err error
	var exists bool
	if err = c.BindJSON(&body); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("请求参数据解析失败")))
		return
	}
	if body.Password == "" || body.Account == "" {
		c.JSON(500, resp.Error(errors.New("参数错误"), resp.Msg("缺少账号,密码")))
		return
	}

	hash := accountutil.Password(body.Password)
	account := model.User{
		Id:       uuid.String(),
		Name:     body.Name,
		Account:  body.Account,
		Password: hash,
	}
	// 检查账号是否存在
	if exists, err = dao.AccountMapper.IsRegister(account); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("注册失败,请联系管理员")))
		return
	}
	if exists {
		c.JSON(500, resp.Error(errors.New("账号已存在"), resp.Msg("账号已存在")))
		return
	}
	// 开始注册账号
	var begin *sql.Tx

	if begin, err = db.DB.Begin(); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("注册失败,请联系管理员")))
		return
	}
	if err = dao.AccountMapper.Register(account, begin); err != nil {
		logs.Error(err.Error())
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("注册失败,请联系管理员")))
		return
	}
	// todo 给用户分配默认组织和默认角色 第一次注册统一存放到根组织下,分配 普通用户 角色
	// 1.分配组织 并设置当前组织为首选项组织
	param := map[string]interface{}{
		"Id":          uuid.String(),
		"UserId":      account.Id,
		"OrgId":       ROOT_ORG_ID,
		"FirstChoice": true,
	}
	if err = dao.OrgMapper.OrgAddUser(param, begin); err != nil {
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("注册失败,请联系管理员")))
		return
	}
	// 2.分配角色 并设置当前角色为 当前组织的首选项角色
	role := model.AuthUserRole{
		Id:          uuid.String(),
		UserId:      account.Id,
		RoleId:      ROOT_ORG_DEFAULT_ROLE,
		OrgId:       ROOT_ORG_ID,
		FirstChoice: true,
	}
	if err = dao.AuthMapper.RegisterAddOrgUserRoleAuth(role, begin); err != nil {
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("注册失败,请联系管理员")))
		return
	}
	// todo 默认注册用户定制化配置(默认授权一部分工具或者路由)
	if err = InitRegisterUser(account, begin); err != nil {
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("注册失败,请联系管理员")))
		return
	}
	// todo 创建用户消息队列
	key := fmt.Sprintf("%s%s", mq_key.Notify, account.Id)
	// 每个用户创建一个任务队列 用于通知消息
	rabbmq.CreateUserNotifyQueue(key)
	begin.Commit()
	c.JSON(200, resp.Success(account, resp.Msg("注册成功")))
}

// Login
// @Summary      用户登录
// @Description  系统用户进行系统登陆
// @Tags         用户接口
// @Accept       json
// @Param        args body  LoginArgs true "登录参数"
// @Produce      json
// @Success      200  {object}  resp.Response{code=int,data=any,msg=string}
// @Failure      500  {object}  resp.Response{code=int,data=any,msg=string}
// @Router       /public/login [post]
func Login(c *gin.Context) {
	var body LoginArgs
	var err error
	var exists bool
	var account model.User
	web.BindJSON(c, &body)
	if body.Account == "" || body.Password == "" {
		c.JSON(500, resp.Error(errors.New("参数错误"), resp.Msg("缺少账号,密码")))
		return
	}
	account.Account = body.Account
	if exists, err = dao.AccountMapper.IsRegister(account); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("注册失败,请联系管理员")))
		return
	}
	if !exists {
		c.JSON(500, resp.Error(errors.New("密码错误"), resp.Msg("密码错误")))
		return
	}
	if account, err = dao.AccountMapper.SelectAccount(map[string]any{
		"Account": body.Account,
	}); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("登录失败,请联系管理员")))
		return
	}
	if !accountutil.VerifyPasswd(account.Password, body.Password) {
		c.JSON(500, resp.Error(errors.New("密码错误"), resp.Msg("密码错误")))
		return
	}
	// 生成 app token
	var tokenStr string
	if tokenStr, err = auth.CreateToken(account); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("登录失败,请联系管理员")))
		return
	}

	data := &model.AppNotify{
		Id:         uuid.String(),
		PubId:      "system",
		SubId:      account.Id,
		Title:      "登录通知",
		MsgType:    1,
		Text:       "成功登录",
		CreateTime: time.Now().Format(time.DateTime),
		UpdateTime: time.Now().Format(time.DateTime),
	}
	rabbmq.Notify(data)
	c.JSON(200, resp.Success(map[string]any{
		"token": tokenStr,
	}, resp.Msg("登录成功")))
}

func NotifyPull(c *gin.Context) {

}

func Notify(c *gin.Context) {
	token := c.MustGet(auth.Key).(*auth.Token)
	var upgrader = websocket.Upgrader{
		Subprotocols: []string{token.Value},
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	var con *websocket.Conn
	var err error
	if con, err = upgrader.Upgrade(c.Writer, c.Request, nil); err != nil {
		logs.Error("upgrade:" + err.Error())
		return
	}
	defer con.Close()
	openMQ(con, token)
}
func openMQ(con *websocket.Conn, token *auth.Token) {
	key := fmt.Sprintf("%s%s", mq_key.Notify, token.Id)
	var err error
	var ch *amqp.Channel
	if ch, err = rabbmq.Client.Channel(); err != nil {
		logs.Error(err.Error())
		return
	}
	defer ch.Close()

	var msgs <-chan amqp.Delivery
	msgs, err = ch.Consume(
		key,   // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		logs.Error(err.Error())
		return
	}

	// 接收消息 并处理消息
	for msg := range msgs {
		bodyLog := fmt.Sprintf("Received a message: %s", msg.Body)
		logs.Info(bodyLog)
		con.WriteMessage(websocket.TextMessage, msg.Body)
	}
}

func AllUser(c *gin.Context) {
	var err error
	var args *PageArgs
	web.ShouldJSON(c, &args)
	if args.Start, args.End, err = pageutils.PageNumber(args.Page, args.PageSize); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("分页参数错误")))
		return
	}
	var users []*model.User
	var count int64
	if users, count, err = dao.SystemMapper.AllUserList(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	c.JSON(200, resp.Success(resp.NewPage(count, users), resp.Msg("查询成功")))
}

// UpdateUserInfo
// @Summary      更新用户信息
// @Description  更新用户信息
// @Tags         用户接口
// @Accept       json
// @Param        args body  UpdateUserInfoArgs true "更新参数"
// @Produce      json
// @Success      200  {object}  resp.Response{code=int,data=any,msg=string}
// @Failure      500  {object}  resp.Response{code=int,data=any,msg=string}
// @Router       /api/user/update [post]
func UpdateUserInfo(c *gin.Context) {
	var err error
	var body *UpdateUserInfoArgs
	var begin *sql.Tx
	token := c.MustGet(auth.Key).(*auth.Token)
	web.BindJSON(c, &body)
	if begin, err = db.DB.Begin(); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("更新失败,请联系管理员")))
		return
	}
	params := make(map[string]any)
	params["Id"] = token.Id
	if body.Name != nil {
		params["Name"] = *body.Name
		if err = dao.AccountMapper.UpdateUserName(params, begin); err != nil {
			begin.Rollback()
			logs.Error(err.Error())
			c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
			return
		}
	}
	if body.Age != nil {
		params["Age"] = *body.Age
		if err = dao.AccountMapper.UpdateUserAge(params, begin); err != nil {
			begin.Rollback()
			logs.Error(err.Error())
			c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
			return
		}
	}
	if body.Gender != nil {
		params["Gender"] = *body.Gender
		if err = dao.AccountMapper.UpdateUserGender(params, begin); err != nil {
			begin.Rollback()
			logs.Error(err.Error())
			c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
			return
		}
	}
	begin.Commit()
	c.JSON(200, resp.Success(nil, resp.Msg("修改成功")))
}

// UpdateAvatar
//
//	@Summary      更新用户头像
//
// @Description  更新用户头像
// @Tags         用户接口
// @Accept       multipart/form-data
// @Param		 avatar	formData	file	true	"用户头像文件"
// @Produce      json
// @Success      200  {object}  resp.Response{code=int,data=any,msg=string}
// @Failure      500  {object}  resp.Response{code=int,data=any,msg=string}
// @Router       /api/user/update/avatar [post]
func UpdateAvatar(c *gin.Context) {
	var err error
	var file *multipart.FileHeader
	var open multipart.File
	token := c.MustGet(auth.Key).(*auth.Token)
	// 单文件
	if file, err = c.FormFile("avatar"); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("上传失败")))
		return
	}
	if open, err = file.Open(); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("上传失败")))
		return
	}
	// 创建存储路径
	name := fmt.Sprintf("%s/avatar/%s", token.Id, file.Filename)
	// 执行推送到对象存储
	if _, err = oss.Tencent.Object.Put(context.Background(), name, open, nil); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("上传失败")))
		return
	}
	full := fmt.Sprintf("%s/%s", config.Evn.App.Tencent.BucketURL, name)
	// 更新数据库
	params := map[string]any{
		"Id":      token.Id,
		"Picture": full,
	}
	if err = dao.AccountMapper.UpdateUserAvatar(params); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	c.JSON(200, resp.Success(nil, resp.Msg("修改成功")))
}

// UpdateOrgRole
// @Summary      设置指定组织的默认角色
// @Description  用户修改指定组织的默认角色
// @Tags         用户接口
// @Accept       json
// @Param        args body  UpdateUserOrgRole true "请求体"
// @Produce      json
// @Success      200  {object}  resp.Response{code=int,data=any,msg=string}
// @Failure      500  {object}  resp.Response{code=int,data=any,msg=string}
// @Router       /api/user/org/update/role [post]
func UpdateOrgRole(c *gin.Context) {
	var err error
	var body *UpdateUserOrgRole
	token := c.MustGet(auth.Key).(*auth.Token)
	web.BindJSON(c, &body)
	var begin *sql.Tx
	if begin, err = db.DB.Begin(); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("修改失败,请联系管理员")))
		return
	}
	params := make(map[string]any)
	params["UserId"] = token.Id
	params["OrgId"] = body.OrgId
	params["RoleId"] = body.OldRoleId
	params["Flag"] = false
	if err = dao.AccountMapper.UpdateUserOrgRole(params, begin); err != nil {
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	params["RoleId"] = body.NewRoleId
	params["Flag"] = true
	if err = dao.AccountMapper.UpdateUserOrgRole(params, begin); err != nil {
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	begin.Commit()
	c.JSON(200, resp.Success(nil, resp.Msg("修改成功")))
}

// UpdateUserOrg
// @Summary      设置用户的默认组织
// @Description  用户修改登录系统的默认组织,用户智能有一个默认组织
// @Tags         用户接口
// @Accept       json
// @Param        args body  UpdateUserOrgArgs true "请求体"
// @Produce      json
// @Success      200  {object}  resp.Response{code=int,data=any,msg=string}
// @Failure      500  {object}  resp.Response{code=int,data=any,msg=string}
// @Router       /api/user/org/update/org [post]
func UpdateUserOrg(c *gin.Context) {
	var err error
	var body *UpdateUserOrgArgs
	token := c.MustGet(auth.Key).(*auth.Token)
	web.BindJSON(c, &body)
	var begin *sql.Tx
	if begin, err = db.DB.Begin(); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("修改失败,请联系管理员")))
		return
	}
	params := make(map[string]any)
	params["UserId"] = token.Id
	params["OrgId"] = body.OldOrgId
	params["Flag"] = false
	if err = dao.AccountMapper.UpdateUserOrg(params, begin); err != nil {
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	params["OrgId"] = body.NewOrgId
	params["Flag"] = true
	if err = dao.AccountMapper.UpdateUserOrg(params, begin); err != nil {
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	begin.Commit()
	c.JSON(200, resp.Success(nil, resp.Msg("修改成功")))
}

func GetSecure(c *gin.Context) {
	var err error
	var user model.User
	token := c.MustGet(auth.Key).(*auth.Token)
	params := map[string]any{
		"Id": token.Id,
	}
	if user, err = dao.AccountMapper.SelectUserById(params); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	data := make(map[string]any)
	data["password"] = ""
	if user.Password != "" {
		data["password"] = "******"
	}
	data["phone"] = user.Phone
	if user.Phone != "" {
		data["phone"] = user.Phone[:3] + "****" + user.Phone[7:]
	}
	data["email"] = user.Email
	if user.Email != "" {
		index := strings.Index(user.Email, "@")
		data["email"] = user.Email[:3] + "****" + user.Email[index-2:]
	}
	c.JSON(200, resp.Success(data, resp.Msg("获取成功")))
}

// PhoneLogin
// @Summary      手机登录
// @Description  用户通过手机号登录系统
// @Tags         登录
// @Accept       json
// @Param        args body  PhoneLoginArgs true "请求体"
// @Produce      json
// @Success      200  {object}  resp.Response{code=int,data=any,msg=string}
// @Failure      500  {object}  resp.Response{code=int,data=any,msg=string}
// @Router       /public/phone [post]
func PhoneLogin(c *gin.Context) {
	var err error
	var body *PhoneLoginArgs
	web.BindJSON(c, &body)
	var value = ""
	if value, err = redisUtil.Get(cache.PhoneLoginKey + body.Phone); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("验证码失效")))
		return
	}
	if value != body.Code {
		c.JSON(500, resp.Error(err, resp.Msg("验证码错误")))
		return
	}
	c.JSON(200, resp.Success(nil, resp.Msg("登录成功")))
}

// PhoneCode
// @Summary      登录验证码
// @Description  手机号获取登录验证码
// @Tags         登录
// @Accept       json
// @Produce      json
// @Success      200  {object}  resp.Response{code=int,data=string,msg=string}
// @Failure      500  {object}  resp.Response{code=int,data=any,msg=string}
// @Router       /public/login_code [get]
func PhoneCode(c *gin.Context) {
	value := rand.Intn(100000)
	phone := c.Query("phone")
	v := strconv.Itoa(value * 10)
	if err := redisUtil.SetEx(cache.PhoneLoginKey+phone, v, cache.PhoneCodeTime); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("发送失败")))
		return
	}
	c.JSON(200, resp.Success(v, resp.Msg("获取成功")))
}

// ForgetCode
// @Summary 	忘记密码验证
// @Description (手机号/邮箱号)重置密码获取验证码
// @Tags 		登录
// @Accept 		json
// @Produces 	json
// @Param 		args body ForGetArgs true "请求体"
// @Router 		/public/forget/code [get]
func ForgetCode(c *gin.Context) {
	var args *ForGetArgs
	web.ShouldJSON(c, &args)
	value := rand.Intn(100000)
	key := ""
	if key = args.Phone; key == "" {
		key = args.Email
	}
	if key == "" {
		c.JSON(500, resp.Error(nil, resp.Msg("发送失败")))
		return
	}
	v := strconv.Itoa(value * 10)
	if err := redisUtil.SetEx(cache.ForGetKey+key, v, cache.PhoneCodeTime); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("发送失败")))
		return
	}
	c.JSON(200, resp.Success(v, resp.Msg("获取成功")))
}

// ForgetCodeCheck
// @Summary 	验证码验证
// @Description (手机号/邮箱号)重置密码获取验证码验证
// @Tags 		登录
// @Accept 		json
// @Produces 	json
// @Param 		args body ForGetArgs true "请求体"
// @Router 		/public/forget/code/check [post]
func ForgetCodeCheck(c *gin.Context) {
	var err error
	var args *ForGetArgs
	web.ShouldJSON(c, &args)
	key := ""
	if key = args.Phone; key == "" {
		key = args.Email
	}
	if key == "" {
		c.JSON(500, resp.Error(nil, resp.Msg("验证失败")))
		return
	}
	var value = ""
	if value, err = redisUtil.Get(cache.ForGetKey + key); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("验证码失效")))
		return
	}
	if value != args.Code {
		c.JSON(500, resp.Error(err, resp.Msg("验证码错误")))
		return
	}
	c.JSON(200, resp.Success(nil, resp.Msg("验证成功")))
}

// ResetPassword
// @Summary 	密码重置
// @Description (手机号/邮箱号)重置密码
// @Tags 		登录
// @Accept 		json
// @Produces 	json
// @Param 		args body ForGetArgs true "请求体"
// @Router 		/public/forget/reset [post]
func ResetPassword(c *gin.Context) {
	var err error
	var args *ForGetArgs
	web.ShouldJSON(c, &args)
	newPassword := accountutil.Password(args.Password)
	params := make(map[string]any)
	params["Password"] = newPassword
	if args.Phone != "" {
		params["Phone"] = args.Phone
		if err = dao.AccountMapper.RestUserPasswordByPhone(params); err != nil {
			c.JSON(500, resp.Error(err, resp.Msg("重置失败")))
			return
		}
	}
	if args.Email != "" {
		params["Email"] = args.Email
		if err = dao.AccountMapper.RestUserPasswordByEmail(params); err != nil {
			c.JSON(500, resp.Error(err, resp.Msg("重置失败")))
			return
		}
	}
	c.JSON(200, resp.Success(nil, resp.Msg("重置成功")))
}

// UpdateUserPassword
// @Summary 	更新用户密码
// @Description 用户修改密码
// @Tags 		用户接口
// @Accept 		json
// @Produces 	json
// @Param 		args body SecureArgs true "请求体"
// @Router 		/api/user/secure/update/password [post]
func UpdateUserPassword(c *gin.Context) {
	var err error
	var body *SecureArgs
	web.BindJSON(c, &body)
	token := c.MustGet(auth.Key).(*auth.Token)
	var user model.User
	if user, err = dao.AccountMapper.SelectUserById(map[string]any{"Id": token.Id}); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	hash := accountutil.Password(body.Password)
	if user.Password != hash {
		c.JSON(500, resp.Error(err, resp.Msg("密码错误")))
		return
	}
	newPassword := accountutil.Password(body.NewPassword)
	if err = dao.AccountMapper.UpdateUserPassword(map[string]any{"Id": token.Id, "Password": newPassword}); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	c.JSON(200, resp.Success(nil, resp.Msg("修改成功")))
}

// GetPhoneSecureCode
// @Summary 	更新手机获取验证码
// @Description 用户手机号,获取验证码接口
// @Tags 		用户接口
// @Accept 		json
// @Produces 	json
// @Router 		/api/user/secure/update/phone/code [get]
func GetPhoneSecureCode(c *gin.Context) {
	value := rand.Intn(100000)
	token := c.MustGet(auth.Key).(*auth.Token)
	v := strconv.Itoa(value * 10)
	if err := redisUtil.SetEx(cache.PhoneSecureKey+token.Id, v, cache.PhoneCodeTime); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("发送失败")))
		return
	}
	c.JSON(200, resp.Success(v, resp.Msg("获取成功")))
}

// CheckPhoneCode
// @Summary 	验证码校验
// @Description 更新用户手机号,验证码校验
// @Tags 		用户接口
// @Accept 		json
// @Produces 	json
// @Param 		args body SecureArgs true "请求体"
// @Router 		/api/user/secure/update/phone/code/check [post]
func CheckPhoneCode(c *gin.Context) {
	var err error
	var body *SecureArgs
	web.BindJSON(c, &body)
	token := c.MustGet(auth.Key).(*auth.Token)
	var value = ""
	if value, err = redisUtil.Get(cache.PhoneSecureKey + token.Id); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("验证码失效")))
		return
	}
	if value != body.Code {
		c.JSON(500, resp.Error(err, resp.Msg("验证码错误")))
		return
	}
	c.JSON(200, resp.Success(nil, resp.Msg("验证码正确")))
}

// UpdateUserPhone
// @Summary 	更新用户手机
// @Description 修改用户手机号
// @Tags 		用户接口
// @Accept 		json
// @Produces 	json
// @Param 		args body SecureArgs true "请求体"
// @Router 		/api/user/secure/update/phone [post]
func UpdateUserPhone(c *gin.Context) {
	var err error
	var body *SecureArgs
	web.BindJSON(c, &body)
	token := c.MustGet(auth.Key).(*auth.Token)
	var check bool
	var value = ""
	// 校验验证码
	if value, err = redisUtil.Get(cache.PhoneSecureKey + token.Id); err != nil {
		c.JSON(500, resp.Error(nil, resp.Msg("验证码失效")))
		return
	}
	// 使用完成之后删除验证码
	defer func(key string) {
		err := redisUtil.Del(key)
		if err != nil {
			logs.Error(err.Error())
		}
	}(cache.PhoneSecureKey + token.Id)

	if value != body.Code {
		c.JSON(500, resp.Error(nil, resp.Msg("验证码错误")))
		return
	}
	params := map[string]any{"Id": token.Id, "Phone": body.Phone}
	if check, err = dao.AccountMapper.CheckUserPhone(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	if check {
		c.JSON(500, resp.Error(err, resp.Msg("手机号已存在")))
		return
	}
	if err = dao.AccountMapper.UpdateUserPhone(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	c.JSON(200, resp.Success(nil, resp.Msg("修改成功")))
}

// UpdateUserEmail
// @Summary 	更新用户邮箱
// @Description 更新用户邮箱
// @Tags 		用户接口
// @Accept 		json
// @Produces 	json
// @Param 		args body SecureArgs true "请求体"
// @Router 		/api/user/secure/update/email [post]
func UpdateUserEmail(c *gin.Context) {
	var err error
	var body *SecureArgs
	web.BindJSON(c, &body)
	token := c.MustGet(auth.Key).(*auth.Token)
	var check bool
	params := map[string]any{"Id": token.Id, "Email": body.Email}
	if check, err = dao.AccountMapper.CheckUserEmail(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	if check {
		c.JSON(500, resp.Error(err, resp.Msg("邮箱已被绑定")))
		return
	}

	// 生成随机验证码
	value := rand.Intn(100000)
	v := strconv.Itoa(value * 10)
	verify := base64.StdEncoding.EncodeToString([]byte(v))
	userId := base64.StdEncoding.EncodeToString([]byte(token.Id))
	email := base64.StdEncoding.EncodeToString([]byte(body.Email))
	urlStr := url.URL{
		Scheme: "http",
		User:   nil,
		Host:   "localhost:5173/#/verify",
	}
	values := url.Values{
		"verify": []string{verify},
		"userId": []string{userId},
		"email":  []string{email},
	}
	urlStr.Path = "/" + base64.StdEncoding.EncodeToString([]byte(values.Encode()))
	// 发送激活链接
	sprintf := fmt.Sprintf("<a href=\"%s\" target=\"_blank\">%s</a>", urlStr.String(), urlStr.String())
	logs.Info(sprintf)
	sprintf = strings.ReplaceAll(sprintf, "%2F%23%2F", "/#/")
	logs.Info(sprintf)
	if err = email163.SendHtml("邮箱绑定验证!", sprintf, body.Email); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("发送失败")))
		return
	}

	if err = redisUtil.SetEx(cache.EmailVerifyKey+token.Id, v, cache.EmailVerifyCodeTime); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("发送失败")))
		return
	}

	c.JSON(200, resp.Success(nil, resp.Msg("发送成功")))
}

// CheckPassword
// @Summary 	验证用户密码
// @Description 验证用户密码
// @Tags 		用户接口
// @Accept 		json
// @Produces 	json
// @Param 		args body PasswordVerify true "请求体"
// @Router 		/api/user/secure/check/password [post]
func CheckPassword(c *gin.Context) {
	var err error
	var body *PasswordVerify
	web.BindJSON(c, &body)
	token := c.MustGet(auth.Key).(*auth.Token)
	var user model.User
	if user, err = dao.AccountMapper.SelectUserById(map[string]any{"Id": token.Id}); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	hash := accountutil.Password(body.Password)
	if user.Password != hash {
		c.JSON(500, resp.Error(err, resp.Msg("密码错误")))
		return
	}
	c.JSON(200, resp.Success(nil, resp.Msg("密码正确")))
}

// CheckEmailVerify
// @Summary 	验证用户密码
// @Description 验证用户密码
// @Tags 		用户接口
// @Accept 		json
// @Produces 	json
// @Param	    verify	path string	true	"验证信息"
// @Router 		/public/secure/email/:verify [post]
func CheckEmailVerify(c *gin.Context) {
	var err error
	var body EmailVerify
	web.ShouldBindUri(c, &body)
	var decodeString []byte
	if decodeString, err = base64.StdEncoding.DecodeString(body.Params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("验证失败"), resp.Code(resp.EmailVerifyErr)))
		return
	}
	// 解码用户绑定验证参数
	query := url.Values{}
	if query, err = url.ParseQuery(string(decodeString)); err != nil {
		return
	}
	// 反编码参数
	var buf []byte
	if buf, err = base64.StdEncoding.DecodeString(query.Get("verify")); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("验证失败"), resp.Code(resp.EmailVerifyErr)))
		return
	}
	verify := string(buf)
	if buf, err = base64.StdEncoding.DecodeString(query.Get("userId")); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("验证失败"), resp.Code(resp.EmailVerifyErr)))
		return
	}
	userId := string(buf)
	if buf, err = base64.StdEncoding.DecodeString(query.Get("email")); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("验证失败"), resp.Code(resp.EmailVerifyErr)))
		return
	}
	email := string(buf)
	// 验证绑定邮箱的有效时间
	get := ""
	if get, err = redisUtil.Get(cache.EmailVerifyKey + userId); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(nil, resp.Msg("验证过期,请重新绑定"), resp.Code(resp.EmailVerifyErr)))
		return
	}
	defer redisUtil.Del(cache.EmailVerifyKey + userId)
	if get != verify {
		logs.Error("验证码不匹配")
		c.JSON(500, resp.Error(err, resp.Msg("验证失败"), resp.Code(resp.EmailVerifyErr)))
		return
	}
	// 二次验证邮箱是否被其他用户绑定
	var check bool
	params := map[string]any{"Id": userId, "Email": email}
	if check, err = dao.AccountMapper.CheckUserEmail(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(nil, resp.Msg("绑定失败"), resp.Code(resp.EmailVerifyErr)))
		return
	}
	if check {
		c.JSON(500, resp.Error(err, resp.Msg("邮箱已被绑定"), resp.Code(resp.EmailVerifyErr)))
		return
	}
	if err = dao.AccountMapper.UpdateUserEmail(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(nil, resp.Msg("绑定失败"), resp.Code(resp.EmailVerifyErr)))
		return
	}
	c.JSON(200, resp.Success(nil, resp.Msg("绑定成功")))
}

// InitRegisterUser
// 初始化注册用户 所有的初始化操作都基于默认root组织下的普通角色
func InitRegisterUser(user model.User, begin *sql.Tx) error {
	var err error
	// 1. 分配 GPT 插件工具
	list := []model.AuthTool{
		{Id: uuid.String(), UserId: user.Id, OrgId: ROOT_ORG_ID, ToolId: GPT_TOOL_ID, RoleId: ROOT_ORG_DEFAULT_ROLE},
	}
	params := map[string]any{
		"list": list,
	}
	if err = dao.AuthMapper.AddOrgUserRoleToolAuth(params, begin); err != nil {
		return err
	}
	// 2. 初始化用户所有的插件配置项
	var templates []model.AppSetting
	if templates, err = setting.GetSettingTemplate(); err != nil {
		return err
	}
	for i := range templates {
		templates[i].Id = uuid.String()
		templates[i].UserId = user.Id
	}
	params = map[string]any{
		"list": templates,
	}
	if err = dao.AccountMapper.AddSetting(params, begin); err != nil {
		return err
	}
	return nil
}

func GetUid(c *gin.Context) {
	c.JSON(200, resp.Success(uuid.String(), resp.Msg("获取成功")))
}
