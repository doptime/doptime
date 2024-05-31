package dynamicdev

import (
	"os"

	"github.com/doptime/doptime/api"
)

type CodePutIn struct {
	FileName   string
	SourceCode string
}
type CodePutOut struct {
	Architect  string
	SourceCode string
}

var APICodePut = api.Api(func(paramIn *CodePutIn) (architectures *CodePutOut, err error) {
	architectures = &CodePutOut{SourceCode: paramIn.SourceCode}
	if architectures.Architect, err = SourceCodeToArchitecture(paramIn.SourceCode); err != nil {
		return nil, err
	}
	//save the code to local file
	if paramIn.SourceCode == "" || paramIn.FileName == "" {
		return architectures, nil
	}
	file, err := os.Create(paramIn.FileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	_, err = file.WriteString(paramIn.SourceCode)
	if err != nil {
		return nil, err
	}
	return architectures, nil
}).Func
