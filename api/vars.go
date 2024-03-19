package api

type InEmpty struct {
}
type InIDPub struct {
	Id  int64 `msgpack:"alias:Jwtid"`
	Pub int64 `msgpack:"alias:Jwtpub"`
}
