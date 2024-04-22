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
	"github.com/jimu-server/oss"
	"github.com/jimu-server/redis/cache"
	"github.com/jimu-server/redis/redisUtil"
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
	if exists, err = AccountMapper.IsRegister(account); err != nil {
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
	if err = AccountMapper.Register(account, begin); err != nil {
		logs.Error(err.Error())
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("注册失败,请联系管理员")))
		return
	}
	// todo 给用户分配默认组织和默认角色 第一次注册统一存放到根组织下,分配普通角色
	// 1.分配组织
	param := map[string]interface{}{
		"Id":          uuid.String(),
		"UserId":      account.Id,
		"OrgId":       "1",
		"FirstChoice": true,
	}
	if err = OrgMapper.OrgAddUser(param, begin); err != nil {
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("注册失败,请联系管理员")))
		return
	}
	role := model.AuthUserRole{
		Id:          uuid.String(),
		UserId:      account.Id,
		RoleId:      "3",
		OrgId:       "1",
		FirstChoice: true,
	}
	if err = AuthMapper.RegisterAddOrgUserRoleAuth(role, begin); err != nil {
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
	if exists, err = AccountMapper.IsRegister(account); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("注册失败,请联系管理员")))
		return
	}
	if !exists {
		c.JSON(500, resp.Error(errors.New("密码错误"), resp.Msg("密码错误")))
		return
	}
	if account, err = AccountMapper.SelectAccount(map[string]any{
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
	if users, count, err = SystemMapper.AllUserList(args); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("查询失败")))
		return
	}
	c.JSON(200, resp.Success(resp.NewPage(count, users), resp.Msg("查询成功")))
}

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
		if err = AccountMapper.UpdateUserName(params, begin); err != nil {
			begin.Rollback()
			logs.Error(err.Error())
			c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
			return
		}
	}
	if body.Age != nil {
		params["Age"] = *body.Age
		if err = AccountMapper.UpdateUserAge(params, begin); err != nil {
			begin.Rollback()
			logs.Error(err.Error())
			c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
			return
		}
	}
	if body.Gender != nil {
		params["Gender"] = *body.Gender
		if err = AccountMapper.UpdateUserGender(params, begin); err != nil {
			begin.Rollback()
			logs.Error(err.Error())
			c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
			return
		}
	}
	begin.Commit()
	c.JSON(200, resp.Success(nil, resp.Msg("修改成功")))
}

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
	if err = AccountMapper.UpdateUserAvatar(params); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	c.JSON(200, resp.Success(nil, resp.Msg("修改成功")))
}

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
	if err = AccountMapper.UpdateUserOrgRole(params, begin); err != nil {
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	params["RoleId"] = body.NewRoleId
	params["Flag"] = true
	if err = AccountMapper.UpdateUserOrgRole(params, begin); err != nil {
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	begin.Commit()
	c.JSON(200, resp.Success(nil, resp.Msg("修改成功")))
}

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
	if err = AccountMapper.UpdateUserOrg(params, begin); err != nil {
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	params["OrgId"] = body.NewOrgId
	params["Flag"] = true
	if err = AccountMapper.UpdateUserOrg(params, begin); err != nil {
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
	if user, err = AccountMapper.SelectUserById(params); err != nil {
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

func ResetPassword(c *gin.Context) {
	var err error
	var args *ForGetArgs
	web.ShouldJSON(c, &args)
	newPassword := accountutil.Password(args.Password)
	params := make(map[string]any)
	params["Password"] = newPassword
	if args.Phone != "" {
		params["Phone"] = args.Phone
		if err = AccountMapper.RestUserPasswordByPhone(params); err != nil {
			c.JSON(500, resp.Error(err, resp.Msg("重置失败")))
			return
		}
	}
	if args.Email != "" {
		params["Email"] = args.Email
		if err = AccountMapper.RestUserPasswordByEmail(params); err != nil {
			c.JSON(500, resp.Error(err, resp.Msg("重置失败")))
			return
		}
	}
	c.JSON(200, resp.Success(nil, resp.Msg("重置成功")))
}

func UpdateUserPassword(c *gin.Context) {
	var err error
	var body *SecureArgs
	web.BindJSON(c, &body)
	token := c.MustGet(auth.Key).(*auth.Token)
	var user model.User
	if user, err = AccountMapper.SelectUserById(map[string]any{"Id": token.Id}); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	hash := accountutil.Password(body.Password)
	if user.Password != hash {
		c.JSON(500, resp.Error(err, resp.Msg("密码错误")))
		return
	}
	newPassword := accountutil.Password(body.NewPassword)
	if err = AccountMapper.UpdateUserPassword(map[string]any{"Id": token.Id, "Password": newPassword}); err != nil {
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	c.JSON(200, resp.Success(nil, resp.Msg("修改成功")))
}

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
	if check, err = AccountMapper.CheckUserPhone(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	if check {
		c.JSON(500, resp.Error(err, resp.Msg("手机号已存在")))
		return
	}
	if err = AccountMapper.UpdateUserPhone(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(err, resp.Msg("修改失败")))
		return
	}
	c.JSON(200, resp.Success(nil, resp.Msg("修改成功")))
}

func UpdateUserEmail(c *gin.Context) {
	var err error
	var body *SecureArgs
	web.BindJSON(c, &body)
	token := c.MustGet(auth.Key).(*auth.Token)
	var check bool
	params := map[string]any{"Id": token.Id, "Email": body.Email}
	if check, err = AccountMapper.CheckUserEmail(params); err != nil {
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

func CheckPassword(c *gin.Context) {
	var err error
	var body *PasswordVerify
	web.BindJSON(c, &body)
	token := c.MustGet(auth.Key).(*auth.Token)
	var user model.User
	if user, err = AccountMapper.SelectUserById(map[string]any{"Id": token.Id}); err != nil {
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
	if check, err = AccountMapper.CheckUserEmail(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(nil, resp.Msg("绑定失败"), resp.Code(resp.EmailVerifyErr)))
		return
	}
	if check {
		c.JSON(500, resp.Error(err, resp.Msg("邮箱已被绑定"), resp.Code(resp.EmailVerifyErr)))
		return
	}
	if err = AccountMapper.UpdateUserEmail(params); err != nil {
		logs.Error(err.Error())
		c.JSON(500, resp.Error(nil, resp.Msg("绑定失败"), resp.Code(resp.EmailVerifyErr)))
		return
	}
	c.JSON(200, resp.Success(nil, resp.Msg("绑定成功")))
}
