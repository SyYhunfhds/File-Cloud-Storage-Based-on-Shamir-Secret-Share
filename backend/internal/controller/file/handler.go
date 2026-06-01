package file

import (
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func BatchUploadHandler(
	uploadDir string,
) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		files := r.GetUploadFiles("items") // 带审计的条目
		if len(files) == 0 {
			r.Response.WriteJson(g.Map{
				"code":    1,
				"message": "请选择上传文件",
			})
			return
		}

		names, err := files.Save(uploadDir, false) //是否 启用随机重命名
		if err != nil {
			r.Response.WriteJson(g.Map{
				"code":    1,
				"message": fmt.Sprintf("文件保存失败: %v", err.Error()),
			})
			return
		}

		r.Response.WriteJson(g.Map{
			"code":    0,
			"message": "上传成功",
			"data": g.Map{
				"count": len(names),
				"files": names,
			},
		})
	}
}
