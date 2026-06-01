package main

import (
	_ "backend/internal/packed"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"backend/internal/cmd"
)

func main() {
	ctx := gctx.GetInitCtx()
	cmd.Main.Run(ctx)
}
