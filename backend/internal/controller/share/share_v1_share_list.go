package share

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"backend/api/share/v1"
)

func (c *ControllerV1) ShareList(ctx context.Context, req *v1.ShareListReq) (res *v1.ShareListRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
