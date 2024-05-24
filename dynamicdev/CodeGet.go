package dynamicdev

import (
	"bufio"
	"os"
	"strings"

	"github.com/doptime/doptime/api"
)

type CodeGetIn struct {
	FileName string
}
type CodeGetOut struct {
	SourceCode string
}

var dirOfProject, _ = os.Getwd()

func ReadInGoFile(filePath string) (content string, err error) {

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var contentBuilder strings.Builder

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		contentBuilder.WriteString(scanner.Text() + "\n")
	}

	if err = scanner.Err(); err != nil {
		return "", err
	}

	return contentBuilder.String(), nil
}

var APICodeGet = api.Api(func(paramIn *CodeGetIn) (result *CodeGetOut, err error) {
	result = &CodeGetOut{}
	if result.SourceCode, err = ReadInGoFile(dirOfProject + paramIn.FileName); err != nil {
		return nil, err
	}
	return result, nil
}).Func
