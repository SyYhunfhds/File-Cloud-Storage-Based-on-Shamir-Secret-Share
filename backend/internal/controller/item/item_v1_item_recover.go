package item

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"backend/api/item/v1"
)

func (c *ControllerV1) ItemRecover(ctx context.Context, req *v1.ItemRecoverReq) (res *v1.ItemRecoverRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
