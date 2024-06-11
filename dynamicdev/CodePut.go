// Description:
// 1st: The client will upload source code in this API.
// 2nd: A file named {FileName}.tmp will be created. And the {FileName}.tmp will compared with the FileName specified In the input parameter using command line code diff <FileName> <FileName>.tmp
// 3rd: this vscode windows Will be focused to allow user to make revision
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
