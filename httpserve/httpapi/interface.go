package httpapi

type ApiInterface interface {
	GetName() string
	CallByMap(_map map[string]interface{}, msgpackNonstruct []byte, jsonpackNostruct []byte) (ret interface{}, err error)
	GetDataSource() string
}
