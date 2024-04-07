-- 组织
drop table if exists app_org;
create table app_org
(
    id          varchar(30) primary key,
    pid         varchar(30),
    name        varchar(30),
    create_time timestamp(0) not null default now()
);

comment on table app_org is '组织表 ';
comment on column app_org.pid is '上级组织 ';
comment on column app_org.name is '用户昵称 ';
comment on column app_org.create_time is '创建时间';


-- 角色
drop table if exists app_role;
create table app_role
(
    id          varchar(30) primary key,
    name        varchar(30),
    role_key    varchar(50)  not null,
    create_time timestamp(0) not null default now()
);

comment on table app_role is '角色 ';
comment on column app_role.name is '角色名称 ';
comment on column app_role.role_key is '角色标识符 ';
comment on column app_role.create_time is '创建时间';


-- 用户分配组织
drop table if exists app_org_user;
create table app_org_user
(
    id           varchar(30) primary key,
    org_id       varchar(30)  not null,
    user_id      varchar(30)  not null,
    first_choice bool         not null default true,
    create_time  timestamp(0) not null default now()
);
comment on table app_org_user is '用户组织关系表 ';
comment on column app_org_user.org_id is '组织id';
comment on column app_org_user.user_id is '用户id ';
comment on column app_org_user.first_choice is '当前组织是否为首选组织';
comment on column app_org_user.create_time is '创建时间';

-- 初始化组织角色
drop table if exists app_org_role;
create table app_org_role
(
    id          varchar(30) primary key,
    org_id      varchar(30)  not null,
    role_id     varchar(50)  not null,
    create_time timestamp(0) not null default now()
);
comment on table app_org_role is '组织角色表';
comment on column app_org_role.org_id is '角色所属组织id';
comment on column app_org_role.role_id is '角色id ';
comment on column app_org_role.create_time is '创建时间';

drop table if exists app_org_user_role;
create table app_org_user_role
(
    id          varchar(30) primary key,
    org_id      varchar(30)  not null,
    user_id     varchar(30)  not null,
    role_id     varchar(50)  not null,
    create_time timestamp(0) not null default now()
);
comment on table app_org_user_role is '组织用户角色分配表';
comment on column app_org_user_role.org_id is '角色所属组织id';
comment on column app_org_user_role.role_id is '角色id ';
comment on column app_org_user_role.user_id is '用户id ';
comment on column app_org_user_role.create_time is '创建时间';

-- 工具栏
drop table if exists app_tool;
create table app_tool
(
    id        varchar(30) primary key,
    name      varchar(100) not null,
    icon      varchar(100) not null,
    component varchar(200) not null,
    path      varchar(200) not null,
    tip       varchar(100) not null,
    position  int          not null
);
comment on table app_tool is '周边工具栏表';
comment on column app_tool.name is '路由名称,并且不能重复';
comment on column app_tool.icon is '图标';
comment on column app_tool.component is '工具对应窗口组件';
comment on column app_tool.path is '工具基础路径 工具栏下的所有路由都应该基于此 /{name}';
comment on column app_tool.tip is '提示语,一般填写工具名称';
comment on column app_tool.position is '工具按钮位置信息 1:左侧 2:右侧';

-- 初始化系统菜单项
drop table if exists app_menu;
create table app_menu
(
    id          varchar(30) primary key,
    pid         int,
    title       varchar(100)          default '',
    name        varchar(100)          default '',
    icon        varchar(100)          default '',
    component   varchar(200)          default '',
    path        varchar(100)          default '',
    remark      varchar(500)          default '',
    status      bool                  default true,
    sort        int,
    tool_id     varchar(30)           default 0,
    create_time timestamp(0) not null default now()
);
comment on table app_menu is '菜单路由表';
comment on column app_menu.pid is '父节点';
comment on column app_menu.name is '菜单标题 ';
comment on column app_menu.icon is '组件图标';
comment on column app_menu.component is '组件名称,组件基于前端根路径的路径信息';
comment on column app_menu.path is '路由路径(注册路由时候的注册路径)';
comment on column app_menu.remark is '备注信息';
comment on column app_menu.status is '菜单启用状态 0:未启用 1:启用';
comment on column app_menu.sort is '排序字段';
comment on column app_menu.tool_id is '菜单所属工具栏';
comment on column app_menu.create_time is '创建时间';


-- 功能表
drop table if exists app_fun;
create table app_fun
(
    id          varchar(30) primary key,
    method      varchar(10),
    name        varchar(100),
    path        varchar(100),
    status      boolean               default true,
    create_time timestamp(0) not null default now()
);

comment on table app_fun is '功能路由表';
comment on column app_fun.method is '接口类型';
comment on column app_fun.name is '功能名称';
comment on column app_fun.path is '功能路径';
comment on column app_fun.status is '菜单启用状态 0:未启用 1:启用';


drop table if exists app_role_menu_auth;
create table app_role_menu_auth
(
    id          varchar(30) primary key,
    menu_id     varchar(30)  not null,
    role_id     varchar(30)  not null,
    create_time timestamp(0) not null default now()
);

drop table if exists app_role_tool_auth;
create table app_role_tool_auth
(
    id          varchar(30) primary key,
    tool_id     varchar(30)  not null,
    role_id     varchar(30)  not null,
    create_time timestamp(0) not null default now()
);

drop table if exists app_role_fun_auth;
create table app_role_fun_auth
(
    id          varchar(30) primary key,
    fun_id      varchar(30)  not null,
    role_id     varchar(30)  not null,
    create_time timestamp(0) not null default now()
);


drop table if exists app_user_role_tool_auth;
create table app_user_role_tool_auth
(
    id          varchar(30) primary key,
    user_id     varchar(30)  not null,
    role_id     varchar(30)  not null,
    org_id      varchar(30)  not null,
    tool_id     varchar(30)  not null,
    create_time timestamp(0) not null default now()
);


drop table if exists app_user_role_menu_auth;
create table app_user_role_menu_auth
(
    id          varchar(30) primary key,
    user_id     varchar(30)  not null,
    role_id     varchar(30)  not null,
    org_id      varchar(30)  not null,
    menu_id     varchar(30)  not null,
    create_time timestamp(0) not null default now()
);



drop table if exists app_user_role_fun_auth;
create table app_user_role_fun_auth
(
    id          varchar(30) primary key,
    user_id     varchar(30)  not null,
    role_id     varchar(30)  not null,
    org_id      varchar(30)  not null,
    fun_id      varchar(30)  not null,
    create_time timestamp(0) not null default now()
);


