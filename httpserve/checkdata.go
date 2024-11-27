package httpserve

const (
	HGET             string = "HGET"
	HSET             string = "HSET"
	HDEL             string = "HDEL"
	HGETALL          string = "HGETALL"
	HMGET            string = "HMGET"
	HKEY             string = "HKEY"
	HKEYS            string = "HKEYS"
	HVALS            string = "HVALS"
	HLEN             string = "HLEN"
	HSTRLEN          string = "HSTRLEN"
	HINCRBY          string = "HINCRBY"
	HINCRBYFLOAT     string = "HINCRBYFLOAT"
	HSETNX           string = "HSETNX"
	HEXISTS          string = "HEXISTS"
	HRANDFIELD       string = "HRANDFIELD"
	HSCAN            string = "HSCAN"
	APPEND           string = "APPEND"
	BITCOUNT         string = "BITCOUNT"
	BITFIELD         string = "BITFIELD"
	BITOP            string = "BITOP"
	BITPOS           string = "BITPOS"
	DECR             string = "DECR"
	DECRBY           string = "DECRBY"
	GET              string = "GET"
	GETBIT           string = "GETBIT"
	GETRANGE         string = "GETRANGE"
	GETSET           string = "GETSET"
	INCR             string = "INCR"
	INCRBY           string = "INCRBY"
	INCRBYFLOAT      string = "INCRBYFLOAT"
	MGET             string = "MGET"
	MSET             string = "MSET"
	MSETNX           string = "MSETNX"
	PSETEX           string = "PSETEX"
	SET              string = "SET"
	SETBIT           string = "SETBIT"
	SETEX            string = "SETEX"
	SETNX            string = "SETNX"
	SETRANGE         string = "SETRANGE"
	STRLEN           string = "STRLEN"
	DEL              string = "DEL"
	DUMP             string = "DUMP"
	EXISTS           string = "EXISTS"
	EXPIRE           string = "EXPIRE"
	EXPIREAT         string = "EXPIREAT"
	KEYS             string = "KEYS"
	MIGRATE          string = "MIGRATE"
	MOVE             string = "MOVE"
	OBJECT           string = "OBJECT"
	PERSIST          string = "PERSIST"
	PEXPIRE          string = "PEXPIRE"
	PEXPIREAT        string = "PEXPIREAT"
	PTTL             string = "PTTL"
	RANDOMKEY        string = "RANDOMKEY"
	RENAME           string = "RENAME"
	RENAMEX          string = "RENAMEX"
	RENAMENX         string = "RENAMENX"
	RESTORE          string = "RESTORE"
	SORT             string = "SORT"
	TOUCH            string = "TOUCH"
	TTL              string = "TTL"
	TYPE             string = "TYPE"
	UNLINK           string = "UNLINK"
	WAIT             string = "WAIT"
	BLPOP            string = "BLPOP"
	BRPOP            string = "BRPOP"
	BRPOPLPUSH       string = "BRPOPLPUSH"
	LINDEX           string = "LINDEX"
	LINSERT          string = "LINSERT"
	LLEN             string = "LLEN"
	XLEN             string = "XLEN"
	LPOP             string = "LPOP"
	LPUSH            string = "LPUSH"
	LPUSHX           string = "LPUSHX"
	LRANGE           string = "LRANGE"
	LREM             string = "LREM"
	LSET             string = "LSET"
	LTRIM            string = "LTRIM"
	RPOP             string = "RPOP"
	RPOPLPUSH        string = "RPOPLPUSH"
	RPUSH            string = "RPUSH"
	RPUSHX           string = "RPUSHX"
	SADD             string = "SADD"
	SCARD            string = "SCARD"
	SCAN             string = "SCAN"
	SDIFF            string = "SDIFF"
	SDIFFSTORE       string = "SDIFFSTORE"
	SINTER           string = "SINTER"
	SINTERSTORE      string = "SINTERSTORE"
	SISMEMBER        string = "SISMEMBER"
	SMEMBERS         string = "SMEMBERS"
	SMOVE            string = "SMOVE"
	SPOP             string = "SPOP"
	SRANDMEMBER      string = "SRANDMEMBER"
	SREM             string = "SREM"
	SSCAN            string = "SSCAN"
	SUNION           string = "SUNION"
	SUNIONSTORE      string = "SUNIONSTORE"
	ZADD             string = "ZADD"
	ZCARD            string = "ZCARD"
	ZCOUNT           string = "ZCOUNT"
	ZINCRBY          string = "ZINCRBY"
	ZINTERSTORE      string = "ZINTERSTORE"
	ZLEXCOUNT        string = "ZLEXCOUNT"
	ZPOPMAX          string = "ZPOPMAX"
	ZPOPMIN          string = "ZPOPMIN"
	ZRANGE           string = "ZRANGE"
	ZRANGEBYLEX      string = "ZRANGEBYLEX"
	ZRANGEBYSCORE    string = "ZRANGEBYSCORE"
	ZRANK            string = "ZRANK"
	ZREM             string = "ZREM"
	ZREMRANGEBYLEX   string = "ZREMRANGEBYLEX"
	ZREMRANGEBYRANK  string = "ZREMRANGEBYRANK"
	ZREMRANGEBYSCORE string = "ZREMRANGEBYSCORE"
	ZREVRANGE        string = "ZREVRANGE"
	ZREVRANGEBYLEX   string = "ZREVRANGEBYLEX"
	ZREVRANGEBYSCORE string = "ZREVRANGEBYSCORE"
	ZREVRANK         string = "ZREVRANK"
	ZSCAN            string = "ZSCAN"
	ZSCORE           string = "ZSCORE"
	ZUNIONSTORE      string = "ZUNIONSTORE"
	XRANGE           string = "XRANGE"
	XRANGEN          string = "XRANGEN"
	XREVRANGE        string = "XREVRANGE"
	XREVRANGEN       string = "XREVRANGEN"
	XREAD            string = "XREAD"
	XADD             string = "XADD"
	XDEL             string = "XDEL"
	TIME             string = "TIME"
)

var DataCmdRequireKey = map[string]bool{
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
	TIME:             false,
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
	XRANGEN:          true,
	XREVRANGE:        true,
}

var DataCmdRequireField = map[string]bool{
	HGET:         true,
	HSET:         true,
	HDEL:         true,
	HINCRBY:      true,
	HINCRBYFLOAT: true,
	HSETNX:       true,
	HEXISTS:      true,
	HSCAN:        false,
	HSTRLEN:      true,
	// 非哈希命令
	APPEND:           false,
	BITCOUNT:         false,
	BITFIELD:         false,
	BITOP:            false,
	BITPOS:           false,
	DECR:             false,
	DECRBY:           false,
	GET:              false,
	GETBIT:           false,
	GETRANGE:         false,
	GETSET:           false,
	INCR:             false,
	INCRBY:           false,
	INCRBYFLOAT:      false,
	MGET:             false,
	MSET:             false,
	MSETNX:           false,
	PSETEX:           false,
	SET:              false,
	SETBIT:           false,
	SETEX:            false,
	SETNX:            false,
	SETRANGE:         false,
	STRLEN:           false,
	DEL:              false,
	DUMP:             false,
	EXISTS:           false,
	EXPIRE:           false,
	EXPIREAT:         false,
	KEYS:             false,
	MIGRATE:          false,
	MOVE:             false,
	OBJECT:           false,
	PERSIST:          false,
	PEXPIRE:          false,
	PEXPIREAT:        false,
	PTTL:             false,
	RANDOMKEY:        false,
	RENAME:           false,
	RENAMENX:         false,
	RESTORE:          false,
	SORT:             false,
	TOUCH:            false,
	TTL:              false,
	TYPE:             false,
	UNLINK:           false,
	WAIT:             false,
	BLPOP:            false,
	BRPOP:            false,
	BRPOPLPUSH:       false,
	LINDEX:           false,
	LINSERT:          false,
	LLEN:             false,
	LPOP:             false,
	LPUSH:            false,
	LPUSHX:           false,
	LRANGE:           false,
	LREM:             false,
	LSET:             false,
	LTRIM:            false,
	RPOP:             false,
	RPOPLPUSH:        false,
	RPUSH:            false,
	RPUSHX:           false,
	SADD:             false,
	SCARD:            false,
	SDIFF:            false,
	SDIFFSTORE:       false,
	SINTER:           false,
	SINTERSTORE:      false,
	SISMEMBER:        false,
	SMEMBERS:         false,
	SMOVE:            false,
	SPOP:             false,
	SRANDMEMBER:      false,
	SREM:             false,
	SSCAN:            false,
	SUNION:           false,
	SUNIONSTORE:      false,
	ZADD:             false,
	ZCARD:            false,
	ZCOUNT:           false,
	ZINCRBY:          false,
	ZINTERSTORE:      false,
	ZLEXCOUNT:        false,
	ZPOPMAX:          false,
	ZPOPMIN:          false,
	ZRANGE:           false,
	ZRANGEBYLEX:      false,
	ZRANGEBYSCORE:    false,
	ZRANK:            false,
	ZREM:             false,
	ZREMRANGEBYLEX:   false,
	ZREMRANGEBYRANK:  false,
	ZREMRANGEBYSCORE: false,
	ZREVRANGE:        false,
	ZREVRANGEBYLEX:   false,
	ZREVRANGEBYSCORE: false,
	ZREVRANK:         false,
	ZSCAN:            false,
	ZSCORE:           false,
	ZUNIONSTORE:      false,
	XRANGEN:          false,
	XREVRANGE:        false,
}
