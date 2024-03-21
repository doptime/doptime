package data

import (
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

func (ctx *Ctx[k, v]) ZAdd(members ...redis.Z) (err error) {
	//MarshalRedisZ
	for i := range members {
		if members[i].Member != nil {
			members[i].Member, _ = msgpack.Marshal(members[i].Member)
		}
	}
	status := ctx.Rds.ZAdd(ctx.Ctx, ctx.Key, members...)
	return status.Err()
}
func (ctx *Ctx[k, v]) ZRem(members ...interface{}) (err error) {
	//msgpack marshal members to slice of bytes
	var bytes = make([][]byte, len(members))
	for i, member := range members {
		if bytes[i], err = msgpack.Marshal(member); err != nil {
			return err
		}
	}
	var redisPipe = ctx.Rds.Pipeline()
	for _, memberBytes := range bytes {
		redisPipe.ZRem(ctx.Ctx, ctx.Key, memberBytes)
	}
	_, err = redisPipe.Exec(ctx.Ctx)

	return err
}
func (ctx *Ctx[k, v]) ZRange(start, stop int64) (members []v, err error) {
	var cmd *redis.StringSliceCmd

	if cmd = ctx.Rds.ZRange(ctx.Ctx, ctx.Key, start, stop); cmd.Err() != nil && cmd.Err() != redis.Nil {
		return nil, cmd.Err()
	}
	return ctx.UnmarshalToSlice(cmd.Val())
}
func (ctx *Ctx[k, v]) ZRangeWithScores(start, stop int64) (members []v, scores []float64, err error) {
	cmd := ctx.Rds.ZRangeWithScores(ctx.Ctx, ctx.Key, start, stop)
	return ctx.UnmarshalRedisZ(cmd.Val())
}
func (ctx *Ctx[k, v]) ZRevRangeWithScores(start, stop int64) (members []v, scores []float64, err error) {
	cmd := ctx.Rds.ZRevRangeWithScores(ctx.Ctx, ctx.Key, start, stop)
	return ctx.UnmarshalRedisZ(cmd.Val())
}
func (ctx *Ctx[k, v]) ZRank(member interface{}) (rank int64, err error) {
	var (
		memberBytes []byte
	)
	//marshal member using msgpack
	if memberBytes, err = msgpack.Marshal(member); err != nil {
		return 0, err
	}
	cmd := ctx.Rds.ZRank(ctx.Ctx, ctx.Key, string(memberBytes))
	return cmd.Val(), cmd.Err()
}
func (ctx *Ctx[k, v]) ZRevRank(member interface{}) (rank int64, err error) {
	var (
		memberBytes []byte
	)
	//marshal member using msgpack
	if memberBytes, err = msgpack.Marshal(member); err != nil {
		return 0, err
	}
	cmd := ctx.Rds.ZRevRank(ctx.Ctx, ctx.Key, string(memberBytes))
	return cmd.Val(), cmd.Err()
}
func (ctx *Ctx[k, v]) ZScore(member interface{}) (score float64, err error) {
	var (
		memberBytes []byte
		cmd         *redis.FloatCmd
	)
	//marshal member using msgpack
	if memberBytes, err = msgpack.Marshal(member); err != nil {
		return 0, err
	}
	if cmd = ctx.Rds.ZScore(ctx.Ctx, ctx.Key, string(memberBytes)); cmd.Err() != nil && cmd.Err() != redis.Nil {
		return 0, err
	} else if cmd.Err() == redis.Nil {
		return 0, nil
	}
	return cmd.Result()
}
func (ctx *Ctx[k, v]) ZCard() (length int64, err error) {
	cmd := ctx.Rds.ZCard(ctx.Ctx, ctx.Key)
	return cmd.Result()
}
func (ctx *Ctx[k, v]) ZCount(min, max string) (length int64, err error) {
	cmd := ctx.Rds.ZCount(ctx.Ctx, ctx.Key, min, max)
	return cmd.Result()
}
func (ctx *Ctx[k, v]) ZRangeByScore(opt *redis.ZRangeBy) (out []v, err error) {
	cmd := ctx.Rds.ZRangeByScore(ctx.Ctx, ctx.Key, opt)

	return ctx.UnmarshalToSlice(cmd.Val())
}
func (ctx *Ctx[k, v]) ZRangeByScoreWithScores(opt *redis.ZRangeBy) (out []v, scores []float64, err error) {
	cmd := ctx.Rds.ZRangeByScoreWithScores(ctx.Ctx, ctx.Key, opt)
	return ctx.UnmarshalRedisZ(cmd.Val())
}
func (ctx *Ctx[k, v]) ZRevRangeByScore(opt *redis.ZRangeBy) (out []v, err error) {
	cmd := ctx.Rds.ZRevRangeByScore(ctx.Ctx, ctx.Key, opt)
	return ctx.UnmarshalToSlice(cmd.Val())
}
func (ctx *Ctx[k, v]) ZRevRange(start, stop int64) (out []v, err error) {
	var cmd *redis.StringSliceCmd

	if cmd = ctx.Rds.ZRevRange(ctx.Ctx, ctx.Key, start, stop); cmd.Err() != nil && cmd.Err() != redis.Nil {
		return nil, cmd.Err()
	}
	return ctx.UnmarshalToSlice(cmd.Val())
}
func (ctx *Ctx[k, v]) ZRevRangeByScoreWithScores(opt *redis.ZRangeBy) (out []v, scores []float64, err error) {
	cmd := ctx.Rds.ZRevRangeByScoreWithScores(ctx.Ctx, ctx.Key, opt)
	return ctx.UnmarshalRedisZ(cmd.Val())
}
func (ctx *Ctx[k, v]) ZRemRangeByRank(start, stop int64) (err error) {
	status := ctx.Rds.ZRemRangeByRank(ctx.Ctx, ctx.Key, start, stop)
	return status.Err()
}
func (ctx *Ctx[k, v]) ZRemRangeByScore(min, max string) (err error) {
	status := ctx.Rds.ZRemRangeByScore(ctx.Ctx, ctx.Key, min, max)
	return status.Err()
}
func (ctx *Ctx[k, v]) ZIncrBy(increment float64, member interface{}) (err error) {
	var (
		memberBytes []byte
	)
	//marshal member using msgpack
	if memberBytes, err = msgpack.Marshal(member); err != nil {
		return err
	}
	status := ctx.Rds.ZIncrBy(ctx.Ctx, ctx.Key, increment, string(memberBytes))
	return status.Err()
}
func (ctx *Ctx[k, v]) ZPopMax(count int64) (out []v, scores []float64, err error) {
	cmd := ctx.Rds.ZPopMax(ctx.Ctx, ctx.Key, count)
	return ctx.UnmarshalRedisZ(cmd.Val())
}
func (ctx *Ctx[k, v]) ZPopMin(count int64) (out []v, scores []float64, err error) {
	cmd := ctx.Rds.ZPopMin(ctx.Ctx, ctx.Key, count)
	return ctx.UnmarshalRedisZ(cmd.Val())
}
func (ctx *Ctx[k, v]) ZLexCount(min, max string) (length int64) {
	cmd := ctx.Rds.ZLexCount(ctx.Ctx, ctx.Key, min, max)
	return cmd.Val()
}
func (ctx *Ctx[k, v]) ZScan(cursor uint64, match string, count int64) ([]string, uint64, error) {
	cmd := ctx.Rds.ZScan(ctx.Ctx, ctx.Key, cursor, match, count)
	return cmd.Result()
}
