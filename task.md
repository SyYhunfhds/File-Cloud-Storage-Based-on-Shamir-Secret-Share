## 项目主题
基于Shamir算法的企业财务审计条目存储系统

## 项目架构
全盘照抄Privy.io，但设法将计算任务转移到后端进行, 其他部分与Privy.io相同

## 技术栈
- 桌面端: Flutter
    - 不考虑网页端开发, 因为前端需要接触本地Vault文件系统
- 后端: GoFrame
- 数据库: Psql (后端) | SQLite/Vault (前端)

## 已经完成的部分
- 【算法实现】Go的SSS GF(2^32-5) 算法实现, 带有PSS特性
    - 路径: `backend\pkg\shamir\v2`
- 【后端API实现】登录链路: 
    物理文件路径: 
    - `backend\api\auth`
    - `backend\internal\controller\auth`
    API路径: 
    - `v1/auth/login` (用户登录)
    - `v1/auth/logout` (用户注销)
    - `v1/user` (用户注册)