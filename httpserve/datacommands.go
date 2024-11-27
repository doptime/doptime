package httpserve

type dataCMD string

const (
	HGET             dataCMD = "HGET"
	HSET             dataCMD = "HSET"
	HDEL             dataCMD = "HDEL"
	HGETALL          dataCMD = "HGETALL"
	HKEY             dataCMD = "HKEY"
	HKEYS            dataCMD = "HKEYS"
	HVALS            dataCMD = "HVALS"
	HLEN             dataCMD = "HLEN"
	HSTRLEN          dataCMD = "HSTRLEN"
	HINCRBY          dataCMD = "HINCRBY"
	HINCRBYFLOAT     dataCMD = "HINCRBYFLOAT"
	HSETNX           dataCMD = "HSETNX"
	HEXISTS          dataCMD = "HEXISTS"
	HSCAN            dataCMD = "HSCAN"
	APPEND           dataCMD = "APPEND"
	BITCOUNT         dataCMD = "BITCOUNT"
	BITFIELD         dataCMD = "BITFIELD"
	BITOP            dataCMD = "BITOP"
	BITPOS           dataCMD = "BITPOS"
	DECR             dataCMD = "DECR"
	DECRBY           dataCMD = "DECRBY"
	GET              dataCMD = "GET"
	GETBIT           dataCMD = "GETBIT"
	GETRANGE         dataCMD = "GETRANGE"
	GETSET           dataCMD = "GETSET"
	INCR             dataCMD = "INCR"
	INCRBY           dataCMD = "INCRBY"
	INCRBYFLOAT      dataCMD = "INCRBYFLOAT"
	MGET             dataCMD = "MGET"
	MSET             dataCMD = "MSET"
	MSETNX           dataCMD = "MSETNX"
	PSETEX           dataCMD = "PSETEX"
	SET              dataCMD = "SET"
	SETBIT           dataCMD = "SETBIT"
	SETEX            dataCMD = "SETEX"
	SETNX            dataCMD = "SETNX"
	SETRANGE         dataCMD = "SETRANGE"
	STRLEN           dataCMD = "STRLEN"
	DEL              dataCMD = "DEL"
	DUMP             dataCMD = "DUMP"
	EXISTS           dataCMD = "EXISTS"
	EXPIRE           dataCMD = "EXPIRE"
	EXPIREAT         dataCMD = "EXPIREAT"
	KEYS             dataCMD = "KEYS"
	MIGRATE          dataCMD = "MIGRATE"
	MOVE             dataCMD = "MOVE"
	OBJECT           dataCMD = "OBJECT"
	PERSIST          dataCMD = "PERSIST"
	PEXPIRE          dataCMD = "PEXPIRE"
	PEXPIREAT        dataCMD = "PEXPIREAT"
	PTTL             dataCMD = "PTTL"
	RANDOMKEY        dataCMD = "RANDOMKEY"
	RENAME           dataCMD = "RENAME"
	RENAMENX         dataCMD = "RENAMENX"
	RESTORE          dataCMD = "RESTORE"
	SORT             dataCMD = "SORT"
	TOUCH            dataCMD = "TOUCH"
	TTL              dataCMD = "TTL"
	TYPE             dataCMD = "TYPE"
	UNLINK           dataCMD = "UNLINK"
	WAIT             dataCMD = "WAIT"
	BLPOP            dataCMD = "BLPOP"
	BRPOP            dataCMD = "BRPOP"
	BRPOPLPUSH       dataCMD = "BRPOPLPUSH"
	LINDEX           dataCMD = "LINDEX"
	LINSERT          dataCMD = "LINSERT"
	LLEN             dataCMD = "LLEN"
	LPOP             dataCMD = "LPOP"
	LPUSH            dataCMD = "LPUSH"
	LPUSHX           dataCMD = "LPUSHX"
	LRANGE           dataCMD = "LRANGE"
	LREM             dataCMD = "LREM"
	LSET             dataCMD = "LSET"
	LTRIM            dataCMD = "LTRIM"
	RPOP             dataCMD = "RPOP"
	RPOPLPUSH        dataCMD = "RPOPLPUSH"
	RPUSH            dataCMD = "RPUSH"
	RPUSHX           dataCMD = "RPUSHX"
	SADD             dataCMD = "SADD"
	SCARD            dataCMD = "SCARD"
	SDIFF            dataCMD = "SDIFF"
	SDIFFSTORE       dataCMD = "SDIFFSTORE"
	SINTER           dataCMD = "SINTER"
	SINTERSTORE      dataCMD = "SINTERSTORE"
	SISMEMBER        dataCMD = "SISMEMBER"
	SMEMBERS         dataCMD = "SMEMBERS"
	SMOVE            dataCMD = "SMOVE"
	SPOP             dataCMD = "SPOP"
	SRANDMEMBER      dataCMD = "SRANDMEMBER"
	SREM             dataCMD = "SREM"
	SSCAN            dataCMD = "SSCAN"
	SUNION           dataCMD = "SUNION"
	SUNIONSTORE      dataCMD = "SUNIONSTORE"
	ZADD             dataCMD = "ZADD"
	ZCARD            dataCMD = "ZCARD"
	ZCOUNT           dataCMD = "ZCOUNT"
	ZINCRBY          dataCMD = "ZINCRBY"
	ZINTERSTORE      dataCMD = "ZINTERSTORE"
	ZLEXCOUNT        dataCMD = "ZLEXCOUNT"
	ZPOPMAX          dataCMD = "ZPOPMAX"
	ZPOPMIN          dataCMD = "ZPOPMIN"
	ZRANGE           dataCMD = "ZRANGE"
	ZRANGEBYLEX      dataCMD = "ZRANGEBYLEX"
	ZRANGEBYSCORE    dataCMD = "ZRANGEBYSCORE"
	ZRANK            dataCMD = "ZRANK"
	ZREM             dataCMD = "ZREM"
	ZREMRANGEBYLEX   dataCMD = "ZREMRANGEBYLEX"
	ZREMRANGEBYRANK  dataCMD = "ZREMRANGEBYRANK"
	ZREMRANGEBYSCORE dataCMD = "ZREMRANGEBYSCORE"
	ZREVRANGE        dataCMD = "ZREVRANGE"
	ZREVRANGEBYLEX   dataCMD = "ZREVRANGEBYLEX"
	ZREVRANGEBYSCORE dataCMD = "ZREVRANGEBYSCORE"
	ZREVRANK         dataCMD = "ZREVRANK"
	ZSCAN            dataCMD = "ZSCAN"
	ZSCORE           dataCMD = "ZSCORE"
	ZUNIONSTORE      dataCMD = "ZUNIONSTORE"
)

var DataCmdShouldHaveKey = map[dataCMD]bool{
	HGET:             true,
	HSET:             true,
	HDEL:             true,
	HGETALL:          true,
	HKEY:             true,
	HKEYS:            true,
	HVALS:            true,
	HLEN:             true,
	HSTRLEN:          true,
	HINCRBY:          true,
	HINCRBYFLOAT:     true,
	HSETNX:           true,
	HEXISTS:          true,
	HSCAN:            true,
	APPEND:           true,
	BITCOUNT:         true,
	BITFIELD:         true,
	BITOP:            true,
	BITPOS:           true,
	DECR:             true,
	DECRBY:           true,
	GET:              true,
	GETBIT:           true,
	GETRANGE:         true,
	GETSET:           true,
	INCR:             true,
	INCRBY:           true,
	INCRBYFLOAT:      true,
	MGET:             true,
	MSET:             true,
	MSETNX:           true,
	PSETEX:           true,
	SET:              true,
	SETBIT:           true,
	SETEX:            true,
	SETNX:            true,
	SETRANGE:         true,
	STRLEN:           true,
	DEL:              true,
	DUMP:             true,
	EXISTS:           true,
	EXPIRE:           true,
	EXPIREAT:         true,
	KEYS:             true,
	MIGRATE:          true,
	MOVE:             true,
	OBJECT:           true,
	PERSIST:          true,
	PEXPIRE:          true,
	PEXPIREAT:        true,
	PTTL:             true,
	RANDOMKEY:        true,
	RENAME:           true,
	RENAMENX:         true,
	RESTORE:          true,
	SORT:             true,
	TOUCH:            true,
	TTL:              true,
	TYPE:             true,
	UNLINK:           true,
	WAIT:             true,
	BLPOP:            true,
	BRPOP:            true,
	BRPOPLPUSH:       true,
	LINDEX:           true,
	LINSERT:          true,
	LLEN:             true,
	LPOP:             true,
	LPUSH:            true,
	LPUSHX:           true,
	LRANGE:           true,
	LREM:             true,
	LSET:             true,
	LTRIM:            true,
	RPOP:             true,
	RPOPLPUSH:        true,
	RPUSH:            true,
	RPUSHX:           true,
	SADD:             true,
	SCARD:            true,
	SDIFF:            true,
	SDIFFSTORE:       true,
	SINTER:           true,
	SINTERSTORE:      true,
	SISMEMBER:        true,
	SMEMBERS:         true,
	SMOVE:            true,
	SPOP:             true,
	SRANDMEMBER:      true,
	SREM:             true,
	SSCAN:            true,
	SUNION:           true,
	SUNIONSTORE:      true,
	ZADD:             true,
	ZCARD:            true,
	ZCOUNT:           true,
	ZINCRBY:          true,
	ZINTERSTORE:      true,
	ZLEXCOUNT:        true,
	ZPOPMAX:          true,
	ZPOPMIN:          true,
	ZRANGE:           true,
	ZRANGEBYLEX:      true,
	ZRANGEBYSCORE:    true,
	ZRANK:            true,
	ZREM:             true,
	ZREMRANGEBYLEX:   true,
	ZREMRANGEBYRANK:  true,
	ZREMRANGEBYSCORE: true,
	ZREVRANGE:        true,
	ZREVRANGEBYLEX:   true,
	ZREVRANGEBYSCORE: true,
	ZREVRANK:         true,
	ZSCAN:            true,
	ZSCORE:           true,
	ZUNIONSTORE:      true,
}
var DataCmdShouldHaveField = map[dataCMD]bool{
	HGET:         true,
	HSET:         true,
	HDEL:         true,
	HINCRBY:      true,
	HINCRBYFLOAT: true,
	HSETNX:       true,
	HEXISTS:      true,
	HSTRLEN:      true,
}
