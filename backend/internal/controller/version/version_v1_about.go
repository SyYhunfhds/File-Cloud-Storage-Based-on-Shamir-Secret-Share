package version

import (
	"context"

	"backend/api/version/v1"
)

func (c *ControllerV1) About(ctx context.Context, req *v1.AboutReq) (res *v1.AboutRes, err error) {
	return c.cacheAbout, nil
}
