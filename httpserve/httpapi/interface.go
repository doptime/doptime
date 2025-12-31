package httpapi

type ApiInterface interface {
	GetName() string
	CallByMap(_map map[string]interface{}) (ret interface{}, err error)
	GetDataSource() string
}
