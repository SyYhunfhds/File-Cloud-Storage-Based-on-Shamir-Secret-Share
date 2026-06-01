package calc

import (
	"context"

	"backend/api/calc/v1"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) ShareCalc(ctx context.Context, req *v1.ShareCalcReq) (res *v1.ShareCalcRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
