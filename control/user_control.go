package control

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jimu-server/common/resp"
	"github.com/jimu-server/db"
	"github.com/jimu-server/middleware/auth"
	"github.com/jimu-server/model"
	"github.com/jimu-server/mq/mq_key"
	"github.com/jimu-server/mq/rabbmq"
	"github.com/jimu-server/util/accountutil"
	"github.com/jimu-server/util/pageutils"
	"github.com/jimu-server/util/uuidutils/uuid"
	"github.com/jimu-server/web"
	amqp "github.com/rabbitmq/amqp091-go"
	"net/http"
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
		begin.Rollback()
		c.JSON(500, resp.Error(err, resp.Msg("注册失败,请联系管理员")))
		return
	}
	begin.Commit()
	key := fmt.Sprintf("%s%s", mq_key.Notify, account.Id)
	// 每个用户创建一个任务队列 用于通知消息
	rabbmq.CreateUserNotifyQueue(key)
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
		SubId:      "1",
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
