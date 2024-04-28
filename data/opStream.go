package data

func (ctx *Ctx[k, v]) XLen() (length int64, err error) {
	return ctx.Rds.XLen(ctx.Ctx, ctx.Key).Result()
}
