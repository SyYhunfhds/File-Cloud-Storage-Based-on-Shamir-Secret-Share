package orm

import (
	"backend/internal/dao"
	"fmt"
	"os"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

var (
	db gdb.DB
)

func init() {
	// 设置数据库配置
	if err := gdb.SetConfig(dbConfig); err != nil {
		panic(err)
	}
	db = g.DB()
	if err := db.PingMaster(); err != nil {
		panic(err)
	}
}

// 初始化资源
func TestMain(m *testing.M) {
	// 初始化database
	// ...

	exitCode := m.Run()

	// 关闭或回收资源
	// ...

	// 退出
	os.Exit(exitCode)
}

// 统计给定文件的有效成员数目
func testCountItemMembers(userId int, itemId int) func(t *testing.T) {
	return func(t *testing.T) {
		if err := dao.ItemMembers.DB().PingMaster(); err != nil {
			t.Fatalf("无法连接数据库: %v", err)
		}

		if itemId <= 0 {
			t.SkipNow() // 跳过这个子测试
			return
		}
		var (
			ctx         = gctx.New()
			memberCount = 0
		)
		/*
			select
			    i.owner_id, im.item_id, count(1) + 1 as count, i.owner_id != 1 as horizon_priv_override
			from public.item_members im
			left join public.items i on i.id = im.item_id
			where status = 1
			group by im.item_id, i.owner_id
			limit 10;
		*/

		// 先查是不是自己的, 是的话就默认有一个成员, 不是的话就检测为水平越权
		type ItemInfo struct {
			OwnerId int `json:"owner_id"`
		}
		var res1 = ItemInfo{}
		if err := dao.Items.Ctx(ctx).Where("id", itemId).Scan(&res1); err != nil {
			t.Fatalf("无法确认是否为用户[%d]所有的文件", userId)
		}
		if res1.OwnerId != userId {
			t.Errorf("文件[%d]实际由用户[%d]所有, 检测到用户[%d]正在尝试非法访问", itemId, res1.OwnerId, userId)
			return
		}
		t.Logf("确认文件[%d]为用户[%d]所有", itemId, userId)

		// 再查询其他能看到这个条目的用户
		memberCount, err := dao.ItemMembers.Ctx(ctx).
			Where("item_id=? AND status=?", itemId, dao.SubmissionApproved).
			Count()
		if err != nil {
			t.Fatalf("无法统计条目[%d]对应的有效审计成员数", itemId)
		}
		memberCount++ // 算上所有者自己
		t.Logf("条目[%d]的审计成员总数(包括所有者自己)为: %d", itemId, memberCount)
	}
}
func TestCountItemMembers(t *testing.T) {
	t.Run("测试数据库连接", testCountItemMembers(0, 0))

	var (
		ItemId = 326
		userId = 2392
	)
	var format = "查询条目ID为: %d 的有效成员数"
	t.Run(fmt.Sprintf(format, ItemId), testCountItemMembers(userId, ItemId))

	{
		ItemId = 326
		userId--
	}
	t.Run(fmt.Sprintf(format, ItemId), testCountItemMembers(userId, ItemId))
}

func TestCountItemMembersEncapsuit(t *testing.T) {
	t.Logf("使用封装好的函数重复上一个测试")

	var (
		ItemId = 326
		userId = 2392
	)
	var format = "查询条目ID为: %d 的有效成员数"
	t.Run(fmt.Sprintf(format, ItemId), func(t *testing.T) {
		ctx := gctx.New()
		count, err := dao.CountMembers(ctx, userId, ItemId)
		if err != nil {
			t.Fatal(err.Error())
		}
		t.Logf("条目[%d]的审计成员总数(包括所有者自己)为: %d", ItemId, count)
	})

	{
		ItemId = 326
		userId--
	}
	t.Run(fmt.Sprintf(format, ItemId), func(t *testing.T) {
		ctx := gctx.New()
		count, err := dao.CountMembers(ctx, userId, ItemId)
		if err != nil {
			t.Fatal(err.Error())
		}
		t.Logf("条目[%d]的审计成员总数(包括所有者自己)为: %d", ItemId, count)
	})
}
