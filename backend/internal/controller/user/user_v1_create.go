package user

import (
	"backend/api/user/v1"
	"backend/internal/dao"
	"backend/internal/model/do"
	"context"
	"math"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/spaolacci/murmur3"
)

// Create
//
// 使用事务, 在创建账号的同时创建职员信息
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	// Request如果有参数无法通过校验会直接在路由层被拦截
	var insertId int64
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		passwdHash, herr := c.hashGen(req.Password)
		if herr != nil {
			return herr
		}

		// 获取最低一级的岗位
		lowestJob, innerErr := dao.GetLowestJob(ctx)
		if innerErr != nil {
			return innerErr
		}

		insertId, innerErr = dao.Users.Ctx(ctx).Data(do.Users{
			Username: req.Username,
			Password: passwdHash,
			Email:    req.Email,
		}).InsertAndGetId() // DAO自动生成Repo
		if innerErr != nil {
			// 实测这里会直接把SQL语句包含在error里返回给前端
			return innerErr
		}

		// 补充计算份额坐标
		_, updatedErr := dao.Users.Ctx(ctx).Data(do.Users{
			// 份额坐标 = Hash(用户ID, gf随机数)
			ShareCoor: murmur3.Sum32WithSeed(gconv.Bytes(insertId), math.MaxUint32),
		}).WherePri(insertId).Update()
		if updatedErr != nil {
			return updatedErr
		}

		insertId, innerErr = dao.Employees.Ctx(ctx).Data(do.Employees{
			UserId: insertId,
			JobId:  lowestJob.Id,
		}).InsertAndGetId() // 后续请求看的是雇员ID不是账号ID

		return nil
	})

	res = &v1.CreateRes{
		Id: insertId,
	}
	return
}
