package dynamicdev

import (
	"bufio"
	"doptime/api"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func keepFunctionDefinitionAndRemoveDetail_SourceCodeToArchitecture(content string) string {
	lines := strings.Split(content, "\n")
	var contentBuilder strings.Builder

	var curlyBrackets []string

	for i, lineString := range lines {
		//remove comment on function definition line
		line := strings.Split(lineString, "//")[0]
		//remove leading and trailing spaces
		line = strings.TrimSpace(line)
		//skip empty line
		l := len(line)
		if l == 0 {
			continue
		}
		funcDefinitionStarting, funcDefinitionEnding := false, false

		//capture the function definition line
		//golang hash strict ast tree, so this evidence is enough
		//seek position of func start
		exceptionCaseOfTypeDefinition := len(curlyBrackets) == 0 && len(line) > 5 && line[:4] == "type"
		if line[l-1] == '{' && !exceptionCaseOfTypeDefinition {
			curlyBrackets = append(curlyBrackets, "{")
			funcDefinitionStarting = len(curlyBrackets) == 1
		}
		//exception case of {}, such as string{} or map[string]string{} ...
		exceptionCaseOfVarInitiation := len(line) > 2 && line[len(line)-2:] == "{}"
		//capture the function definition end
		if L2plusFuncEnd, L1FunEnd := line[l-1] == '}', line[0] == '}'; (L2plusFuncEnd || L1FunEnd) && len(curlyBrackets) > 0 && !exceptionCaseOfVarInitiation {
			//pop the last element
			if curlyBrackets[len(curlyBrackets)-1] == "{" {
				curlyBrackets = curlyBrackets[:len(curlyBrackets)-1]
			}
			funcDefinitionEnding = len(curlyBrackets) == 0
		}

		if inFunctionDefinition := len(curlyBrackets) > 0; funcDefinitionEnding || funcDefinitionStarting || !inFunctionDefinition {
			contentBuilder.WriteString(fmt.Sprintf("%d:%s\n", i+1, lineString))

		}

	}

	return strings.Trim(contentBuilder.String(), "\n")
}

func removeStandardLibraryPackages_SourceCodeToArchitecture(fileContentWithName string) (content string, err error) {
	var contentBuilder strings.Builder
	var importStatements []string
	var importStarted bool = false
	for _, line := range strings.Split(fileContentWithName, "\n") {
		seqWithLinestring := strings.SplitN(line, ":", 2)
		_, lineString := seqWithLinestring[0], seqWithLinestring[1]

		if strings.Contains(lineString, "import (") {
			importStarted = true
		} else if importStarted && strings.Contains(lineString, ")") {
			importStatements = append(importStatements, line)
			importStarted = false
			for i := len(importStatements) - 2; i >= 1; i-- {
				//remove the import statements if not contains "."
				if !strings.Contains(importStatements[i], ".") {
					importStatements = append(importStatements[:i], importStatements[i+1:]...)
				}
			}
			//if lefts no import statements, skip the import block
			if len(importStatements) <= 2 {
				importStatements = []string{}
			}
			//append left import statement to contentBuilder
			for _, importStatement := range importStatements {
				contentBuilder.WriteString(importStatement + "\n")
			}
			continue
		}

		if importStarted {
			importStatements = append(importStatements, line)
		} else {
			contentBuilder.WriteString(line + "\n")
		}
	}

	return contentBuilder.String(), nil

}

type GetProjectArchitectureInfoIn struct {
	SouceCodeFilesToFocus string
}
type GetProjectArchitectureInfoOut map[string]string

var APIGetProjectArchitectureInfo = api.Api(func(packInfo *GetProjectArchitectureInfoIn) (architects GetProjectArchitectureInfoOut, err error) {

	ReadInGoFile := func(filePath string) (content string, err error) {

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
	architects = map[string]string{}

	//get bin path as dirPath
	_, binPath, _, _ := runtime.Caller(0)
	binPath = filepath.Dir(binPath) + "/."

	// walkDir recursively walks through a directory and processes all .go files
	filepath.Walk(binPath, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}
		var processedPage string
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			page, err := ReadInGoFile(path)
			if err != nil {
				return err
			}

			// 先确认语法树是否正确，如果正确再进行替换
			if _, err = parser.ParseFile(token.NewFileSet(), "", page, parser.ParseComments); err != nil {
				return err
			}
			if len(packInfo.SouceCodeFilesToFocus) > 0 && strings.Contains(strings.ToLower(path), strings.ToLower(packInfo.SouceCodeFilesToFocus)) {
				var contentBuilder strings.Builder
				for lineNum, line := range strings.Split(page, "\n") {
					contentBuilder.WriteString(fmt.Sprintf("%d:%s\n", lineNum+1, line))
				}
				processedPage = contentBuilder.String()

			} else {
				processedPage = keepFunctionDefinitionAndRemoveDetail_SourceCodeToArchitecture(page)
				if processedPage, err = removeStandardLibraryPackages_SourceCodeToArchitecture(processedPage); err != nil {
					return err
				}
			}
			fmt.Println(path, processedPage)
			architects[path] = processedPage
		}
		return nil
	})
	return nil, nil
}).Func
