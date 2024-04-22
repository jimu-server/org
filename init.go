package user

import (
	"embed"
	"github.com/jimu-server/db"
	"github.com/jimu-server/middleware/auth"
	"github.com/jimu-server/org/control"
	"github.com/jimu-server/web"
)

//go:embed mapper/file/*.xml
var mapperFile embed.FS

func init() {
	db.GoBatis.LoadByRootPath("mapper", mapperFile)
	db.GoBatis.ScanMappers(control.AccountMapper)
	db.GoBatis.ScanMappers(control.ToolMapper)
	db.GoBatis.ScanMappers(control.OrgMapper)
	db.GoBatis.ScanMappers(control.RoleMapper)
	db.GoBatis.ScanMappers(control.FunMapper)
	db.GoBatis.ScanMappers(control.AuthMapper)
	db.GoBatis.ScanMappers(control.DefaultInfoMapper)
	db.GoBatis.ScanMappers(control.MenuMapper)
	db.GoBatis.ScanMappers(control.SystemMapper)

	web.Engine.GET("/menu", control.AllMenu)

	api := web.Engine.Group("/public")
	api.POST("/register", control.Register)                     // 用户注册接口
	api.POST("/login", control.Login)                           // 密码登录
	api.POST("/phone", control.PhoneLogin)                      // 手机登录
	api.GET("/login_code", control.PhoneCode)                   // 手机号获取登录验证码
	api.POST("/secure/email/:verify", control.CheckEmailVerify) // 验证用户邮箱更新

	api.GET("/forget/code", control.ForgetCode)             // (手机号/邮箱号)重置密码获取验证码
	api.POST("/forget/code/check", control.ForgetCodeCheck) // (手机号/邮箱号)重置密码获取验证码验证
	api.POST("/forget/reset", control.ResetPassword)        // (手机号/邮箱号)重置密码

	api = web.Engine.Group("/api", auth.Authorization())

	api.GET("/dictionary", control.Dictionary)
	api.POST("/org/create", control.CreateOrg)                             // 创建组织
	api.POST("/org/delete", control.DeleteOrg)                             // 删除组织
	api.POST("/org/update", control.UpdateOrgInfo)                         // 更新组织信息
	api.GET("/org/list", control.GetOrg)                                   // 获取组织列表
	api.GET("/org/default/role", control.GetOrg)                           // 获取组织列表
	api.GET("/org/user/list", control.GetOrgUserList)                      // 获取组织下所有的用户列表
	api.GET("/org/role/list", control.GetOrgRoleList)                      // 获取组织下所有的角色列表
	api.GET("/org/user/role", control.GetOrgUserRoleList)                  // 获取组织指定用户的所有已授权角色列表
	api.GET("/org/user/tool", control.GetOrgUserAuthTool)                  // 获取组织指定用户的所有已授权工具列表
	api.GET("/org/user/tool/router", control.GetOrgUserAuthToolRouter)     // 获取组织指定用户的所有已授权工具下的所有路由树
	api.POST("/org/role/auth", control.OrgUserRoleAuth)                    // 给组织的用户授权角色
	api.POST("/org/role/auth/tool", control.OrgUserRoleToolAuth)           // 给组织用户的角色授权工具
	api.POST("/org/role/auth/tool/route", control.OrgUserRoleToolRoleAuth) // 给组织用户的角色的工具授权路由
	api.POST("/org/role/create", control.CreateRole)                       // 给组织创建角色

	api.POST("/role/delete", control.DeleteRole)                      // 删除角色
	api.POST("/role/update", control.UpdateRoleInfo)                  // 更新角色信息
	api.GET("/role/list", control.GetRole)                            // 获取角色列表
	api.GET("/role/tool/list", control.OrgRoleToolList)               // 获取角色已授权的工具列表
	api.GET("/role/tool/router/tree", control.OrgRoleToolRouterList)  // 获取角色已授权的工具栏对应的路由列表
	api.POST("/role/tool/auth", control.OrgRoleToolAuth)              // 对角色进行工具栏授权
	api.POST("/role/tool/router/auth", control.OrgRoleToolRouterAuth) // 对角色进行工具栏对应路由授权

	api.POST("/tool/create", control.CreateTool)            // 创建工具
	api.POST("/tool/delete", control.DeleteTool)            // 删除工具
	api.GET("/tool/list", control.GetTool)                  // 获取工具列表
	api.GET("/tool/router/list", control.GetToolRouterList) // 获取工具路由列表

	api.POST("/fun/create", control.CreateFun) // 创建功能
	api.POST("/fun/delete", control.DeleteFun) // 删除功能
	api.GET("/fun/list", control.GetFun)       // 获取功能列表

	api.GET("/user/auth/menu", control.GetAuthMenu)                          // 获取当前用户的已授权菜单
	api.GET("/user/auth/tool", control.GetAuthTool)                          // 获取当前用户的已授权工具栏
	api.GET("/user/auth/tool/menu", control.GetAuthToolMenu)                 // 获取当前用户的已授权工具栏的菜单路由
	api.GET("/user/auth/tool/menu/child", control.GetAuthToolMenuChild)      // 获取当前用户的已授权工具栏的菜单路由的子路由
	api.GET("/user/default/info", control.GetUserDefaultInfo)                // 获取当前用户的默认组织,和默认组织的默认角色
	api.GET("/user/default/org/role", control.GetOrgDefaultRole)             // 获取当前用户 对应组织的默认角色
	api.GET("/user/info", control.UserInfo)                                  // 获取当前用户信息
	api.GET("/user/all_auth_route", control.UserAuthAllRoute)                // 获取用户当前组织的当前角色所有已授权的前端路由
	api.GET("/user/org/list", control.UserJoinOrgList)                       // 获取用户当前所有已加入的组织
	api.GET("/user/org/listTree", control.UserJoinOrgTreeList)               // 获取用户当前所有已加入的组织及其所有下属组织树形结构
	api.GET("/user/org/list/role", control.UserJoinOrgRoleList)              // 获取用户当前已加入的组织下所有的角色
	api.GET("/user/all", control.AllUser)                                    // 获取系统所有用户
	api.POST("/user/update", control.UpdateUserInfo)                         // 更新用户信息
	api.POST("/user/update/avatar", control.UpdateAvatar)                    // 更新用户头像
	api.POST("/user/org/update/role", control.UpdateOrgRole)                 // 设置指定组织的默认角色
	api.POST("/user/org/update/org", control.UpdateUserOrg)                  // 设置用户的默认组织
	api.GET("/user/secure", control.GetSecure)                               // 获取用户账号安全相关信息
	api.POST("/user/secure/update/password", control.UpdateUserPassword)     // 更新用户密码
	api.GET("/user/secure/update/phone/code", control.GetPhoneSecureCode)    // 更新手机获取验证码
	api.POST("/user/secure/update/phone/code/check", control.CheckPhoneCode) // 更新用户手机号,验证码校验
	api.POST("/user/secure/update/phone", control.UpdateUserPhone)           // 更新用户手机
	api.POST("/user/secure/update/email", control.UpdateUserEmail)           // 更新用户邮箱
	api.POST("/user/secure/check/password", control.CheckPassword)           // 验证用户密码
}
