package dynamicdev

import (
	"os"

	"github.com/doptime/doptime/api"
)

type CodeDelIn struct {
	FileName string
}
type CodeDelOut string

var APICodeDel = api.Api(func(paramIn *CodeDelIn) (result CodeDelOut, err error) {
	if err = os.Remove(dirOfProject + paramIn.FileName); err != nil {
		return "false", err
	}
	return "true", nil
}).Func
