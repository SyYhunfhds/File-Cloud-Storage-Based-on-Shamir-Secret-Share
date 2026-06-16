-- 设置搜索路径
SET search_path TO schema_name, public;

create table if not exists public.users (
    id serial primary key,
    share_coor int8 null,
    username varchar(255) unique not null,
    password varchar(255) not null,
    email varchar(255) unique not null,
    created_at timestamp default now(),
    updated_at timestamp default now()
);

alter table public.users add column share_coor int8 null;
-- 存储用户份额的坐标

create table if not exists public.jobs (
    id serial primary key,
    role_privilege int not null, -- 为了实现方便, 角色权限本身将直接映射到Shamir秘密份额数上
    job varchar(32) unique not null,
    created_at timestamp default now(),
    updated_at timestamp default now()
);
-- 创建岗位到角色权限的映射表

-- 预设岗位名称到岗位权重的映射（数字越大权重越高）
insert into
    public.jobs (role_privilege, job)
values (1, '普通员工'),
    (2, '部门主管'),
    (3, '部门经理'),
    (4, '总监'),
    (5, '副总裁'),
    (6, 'CEO') on conflict (job) do nothing;
-- conflict: 当job字段在插入时触发唯一索引检查, 可以跳过冲突的行继续添加其他行

-- drop table if exists public.employees;
create table if not exists public.employees (
    id serial primary key,
    user_id int unique references users (id) on delete cascade on update cascade,
    job_id int references jobs (id) on delete cascade on update cascade,
    created_at timestamp default now(),
    updated_at timestamp default now()
);
-- 修改employees表的行和列

-- 条目表 items
create table if not exists public.items (
    id serial primary key,
    filename varchar(255) not null, -- (原始)文件名
    savename varchar(255) unique not null, -- 保存名; 唯一标识符
    owner_id int references users (id) on delete cascade on update cascade, -- 文件所有者ID
    uploader_id int references users (id) on delete cascade on update cascade, -- 上传者ID, 默认和owner_id一样
    minimum_privilege int not null, -- 要看到这个条目所需的最低权限, 默认和用户的权限一样
    is_public bool not null default false, -- 是否公开/是否可被公开搜索到
    uploaded_at timestamp default now(), -- 上传时间
    changed_at timestamp default now(), -- 修改时间
    deleted_at timestamp default null -- 删除时间
);

-- 记得给owner_id补一个索引
create index items_owner_id_index on items (owner_id);
-- 启用items表的RLS
alter table public.items enable row level security;

create policy can_access_only_it_belongs_to_user on public.items
for all
to psql
using(
    nullif(current_setting('app.current_user_id', true), '') is null -- 没更新的话这个就为空, 就先放行
    or
    (items.owner_id = current_setting('app.current_user_id', true)::int)
    );

-- 份额表 shares
-- 一个更好的版本
create table if not exists public.shares (
    id serial primary key,
    item_id int references items (id) on delete cascade on update cascade,
    user_id int null references users (id) on delete cascade on update cascade, -- 为NULL则表示是项目级的
    owner_id int null references users (id) on delete cascade on update cascade, -- 辅助字段, 减少连表查询
    owner varchar(255) null, -- 辅助字段, 减少连表查询

    share_type varchar(10) not null, -- 份额类型: device / auth / recovery
    share_base64 text not null, -- 份额内容 Base64编码
    code_hash text null, -- 用于RecoveryShare的恢复码哈希值
    version int default 1,
    status varchar(10) default 'active', -- 份额状态: active / expired / revoked

    expire_at timestamp null default now() + interval '1 hour', -- 份额过期时间; 为了向前兼容, 所以这个字段可以为空
    created_at timestamp default now(),
    updated_at timestamp default now()
);
-- 临时加一列
alter table public.shares add column expire_at timestamp null default now() + interval '1 hour';

create unique index unique_share_status on public.shares (item_id, user_id, share_type, status);
create unique index unique_user_type on public.shares (item_id, user_id, share_type);

create index shares_v2_item_id on public.shares (item_id);

create index shares_v2_user_id on public.shares (user_id);

-- tasks表, 持久化所有加密或解密任务, 用于后续审计
create table if not exists public.tasks (
    id serial primary key,
    filename varchar(255) not null, -- 处理对象
    type int not null, -- 任务类型 --TODO: 优化为枚举类型
    is_succeed boolean, -- 任务是否成功
    failed_for_why text, -- 任务失败原因
    created_at timestamp default now(),
    -- started_at timestamp,
    finished_at timestamp
    -- failed_at timestamp
);

drop table public.tasks cascade;

-- create type audit_status as enum ('rejected', 'passed', 'to_examine');
-- item_members 条目成员表, 用于管理条目分享情况
create table if not exists public.item_members (
    id serial primary key,
    item_id int references items (id) on delete cascade on update cascade,
    member_id int references users (id) on delete cascade on update cascade,
    role varchar(20) not null default 'member', -- owner / member
    status smallint not null default 2, -- 0 已移除; 1: 正常; 2: 待审批
    joined_at timestamp, -- 加入时间
    can_download bool default false, -- 是否可以下载
    updated_at timestamp default now(),
    created_at timestamp default now()
);