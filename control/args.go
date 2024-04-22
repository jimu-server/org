package control

type RegisterArgs struct {
	// 昵称
	Name string `json:"name"`
	// 账号
	Account string `json:"account"`
	// 密码
	Password string `json:"password"`
	// 性别
	Gender int `json:"gender"`
	// 年龄
	Age int `json:"age"`
}

type LoginArgs struct {
	// 账号
	Account string `json:"account"`
	// 密码
	Password string `json:"password"`
}

type Search struct {
	Value string `json:"value"`
}

type ListArgs struct {
	Id       string `form:"id" json:"id"`
	Page     int    `form:"page" json:"page" binding:"required,gte=1"`
	PageSize int    `form:"pageSize" json:"pageSize" binding:"required,gte=5"`
	Start    int    `form:"start" json:"start"`
	End      int    `form:"end" json:"end"`
}

type OrgUserListArgs struct {
	OrgId    string `form:"orgId" json:"orgId"`
	Page     int    `form:"page" json:"page" binding:"required,gte=1"`
	PageSize int    `form:"pageSize" json:"pageSize" binding:"required,gte=5"`
	Start    int    `form:"start" json:"start"`
	End      int    `form:"end" json:"end"`
}

type OrgRoleListArgs struct {
	OrgId string `form:"orgId" json:"orgId"`
	PageArgs
}

type PageArgs struct {
	Page     int `form:"page" json:"page" binding:"required,gte=1"`
	PageSize int `form:"pageSize" json:"pageSize" binding:"required,gte=5"`
	Start    int `form:"start" json:"start"`
	End      int `form:"end" json:"end"`
}

type UpdateOrg struct {
	Id   string `form:"id" json:"id" binding:"required"`
	Pid  string `form:"pid" json:"pid"binding:"required"`
	Name string `form:"name" json:"name"binding:"required,min=2,max=10"`
}

type UpdateRole struct {
	Id   string `form:"id" json:"id" binding:"required"`
	Name string `form:"name" json:"name"binding:"required,min=2,max=10"`
}

type DelArgs struct {
	List []string `form:"list" json:"list" binding:"required"`
}

type AuthArgs struct {
	OrgId  string   `form:"orgId" json:"orgId" binding:"required"`
	UserId string   `form:"userId" json:"userId" binding:"required"`
	RoleId string   `form:"roleId" json:"roleId"`
	ToolId string   `form:"toolId" json:"toolId"`
	Auths  []string `form:"auths" json:"auths" binding:"required"`
	UnAuth []string `form:"unAuth" json:"unAuth" binding:"required"`
}

type RoleAuthQuery struct {
	OrgId  string `form:"orgId" json:"orgId" binding:"required"`
	RoleId string `form:"roleId" json:"roleId" binding:"required"`
	ToolId string `form:"toolId" json:"toolId"`
}

type RoleAuthArgs struct {
	OrgId  string   `form:"orgId" json:"orgId" binding:"required"`
	RoleId string   `form:"roleId" json:"roleId"`
	ToolId string   `form:"toolId" json:"toolId"`
	Auths  []string `form:"auths" json:"auths" binding:"required"`
	UnAuth []string `form:"unAuth" json:"unAuth" binding:"required"`
}

type CreateOrgRole struct {
	Id      string
	OrgId   string `form:"orgId" json:"orgId" binding:"required"`
	Name    string `form:"name" json:"name" binding:"required,min=1,max=10"`
	RoleKey string `form:"roleKey" json:"roleKey" binding:"required,min=1,max=10"`
	RoleId  string
}

type ToolRouterArgs struct {
	Pid    string `form:"pid" json:"pid"`
	ToolId string `form:"toolId" json:"toolId"`
	PageArgs
}

type UpdateUserInfoArgs struct {
	Name   *string `form:"name" json:"name"`
	Gender *int    `form:"gender" json:"gender"`
	Age    *int    `form:"age" json:"age"`
}

type UpdateUserOrgRole struct {
	// 老的默认角色id
	OldRoleId string `form:"oldRoleId" json:"oldRoleId" binding:"required"`
	// 新的默认角色id
	NewRoleId string `form:"newRoleId" json:"newRoleId" binding:"required"`
	// 变更默认角色的组织id
	OrgId string `form:"orgId" json:"orgId" binding:"required"`
}

type UpdateUserOrgArgs struct {
	OldOrgId string `form:"oldOrgId" json:"oldOrgId" binding:"required"`
	NewOrgId string `form:"newOrgId" json:"newOrgId" binding:"required"`
}

type UpdateUserPasswordArgs struct {
	Password string `form:"password" json:"password" binding:"required,min=6,max=20"`
}

type PhoneLoginArgs struct {
	Phone string `form:"phone" json:"phone" binding:"required"`
	Code  string `form:"code" json:"code" binding:"required"`
}

type SecureArgs struct {
	NewPassword string `form:"newPassword" json:"newPassword""`
	Password    string `form:"password" json:"password" `
	Code        string `form:"code" json:"code"`
	Phone       string `form:"phone" json:"phone" `
	Email       string `form:"email" json:"email" `
}

type EmailVerify struct {
	Params string `uri:"verify" binding:"required"`
}

type PasswordVerify struct {
	Password string `form:"password" json:"password"`
}

type ForGetArgs struct {
	Phone    string `form:"phone" json:"phone"`
	Code     string `form:"code" json:"code"`
	Email    string `form:"email" json:"email"`
	Password string `form:"password" json:"password"`
}
