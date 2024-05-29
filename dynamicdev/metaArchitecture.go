package dynamicdev

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/doptime/doptime/api"
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
		//exception case of {}, such as string{} or map[string]string{} ... , return mytype{ a: 1, b: 2}
		exceptionCaseOfVarInitiation := len(line) > 2 && line[len(line)-1:] == "}"
		for i = len(line) - 2; i >= 1 && exceptionCaseOfVarInitiation; i-- {
			exceptionCaseOfVarInitiation = line[i] != '{'
		}

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

func SourceCodeToArchitecture(sourceCode string) (architecture string, err error) {
	processedPage := keepFunctionDefinitionAndRemoveDetail_SourceCodeToArchitecture(sourceCode)
	if processedPage, err = removeStandardLibraryPackages_SourceCodeToArchitecture(processedPage); err != nil {
		return "", err
	}
	return processedPage, nil
}

type GetProjectArchitectureInfoIn struct {
	//default is current dir
	ProjectDir        string
	SkippedDirs       []string
	IncludedSurffixes []string
}
type RelativeFileName string
type GetProjectArchitectureInfoOut map[RelativeFileName]string

var APIGetProjectArchitectureInfo = api.Api(func(packInfo *GetProjectArchitectureInfoIn) (architectures GetProjectArchitectureInfoOut, err error) {

	architectures = map[RelativeFileName]string{}
	var surffixType = map[string]string{".go": "go", ".js": "js", ".ts": "js", ".vue": "js", ".jsx": "js", ".tsx": "js", ".html": "text", ".md": "text", ".json": "text", ".mdx": "text", ".toml": "text", ".txt": "text", "yaml": "text"}
	if packInfo.IncludedSurffixes != nil {
		for _, surffix := range packInfo.IncludedSurffixes {
			surffixType[surffix] = "text"
		}
	}

	//get bin path as dirPath
	// _, binPath, _, _ := runtime.产品经理Caller(0)
	// binPath = filepath.Dir(binPath) + "/."
	dir := dirOfProject
	if len(packInfo.ProjectDir) > 0 {
		dir = packInfo.ProjectDir
	}
	var skipDirs = map[string]bool{}
	for _, skippedDir := range packInfo.SkippedDirs {
		skipDirs[skippedDir] = true
	}
	// walkDir recursively walks through a directory and processes all .go files
	filepath.WalkDir(dir+"/.", func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		doctype, typeExisted := surffixType[filepath.Ext(path)]
		if !typeExisted {
			return nil
		}

		page, _ := ReadInFile(path)

		corrupted := false
		if doctype == "go" {
			// 先确认语法树是否正确，如果正确再进行替换
			_, err = parser.ParseFile(token.NewFileSet(), "", page, parser.ParseComments)
			corrupted = err != nil
		}

		fileName := path[len(dir):]
		architectures[RelativeFileName(fileName)] = page
		if (doctype == "go" || doctype == "js") && !corrupted {
			architectures[RelativeFileName(fileName)], _ = SourceCodeToArchitecture(page)
		}
		return nil
	})

	return architectures, nil
}).Func
