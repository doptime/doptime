package httpserve

import (
	"errors"
	"strconv"

	"github.com/doptime/doptime/data"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

var ErrBadCommand = errors.New("error bad command")

func (svcCtx *HttpContext) PostHandler() (ret interface{}, err error) {

	//db := &data.Ctx{Ctx: svcCtx.Ctx, Rds: config.Rds, Key: svcCtx.Key}
	db := data.New[interface{}, interface{}](data.Option.WithKey(svcCtx.Key).WithRds(svcCtx.RedisDataSource))

	//service name is stored in svcCtx.Key
	switch svcCtx.Cmd {
	case "ZADD":
		var Score float64
		var obj interface{}
		if Score, err = strconv.ParseFloat(svcCtx.Req.FormValue("Score"), 64); err != nil {
			return "false", errors.New("parameter Score shoule be float")
		}
		//unmarshal msgpack
		if MsgPack := svcCtx.MsgpackBodyBytes(); len(MsgPack) == 0 {
			return "false", errors.New("missing MsgPack content")
		} else if err = msgpack.Unmarshal(MsgPack, &obj); err != nil {
			return "false", err
		}
		if err = db.ZAdd(redis.Z{Score: Score, Member: obj}); err != nil {
			return "false", err
		}
		return "true", nil
	default:
		err = ErrBadCommand
	}

	return ret, err
}
