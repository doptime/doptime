package dynamicdev

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/doptime/doptime/api"
)

type StructInfo struct {
	StructName            string
	FieldsWithNameTypeTag []string
}
type RedisData struct {
	RedisKeyType string
	KeyName      string
	KeyType      string
	ValType      string
}

type ArchitechInfoOfAPackage struct {
	StructInfos     []*StructInfo
	FuncProtos      []string
	TypeProtos      []string
	RedisDataProtos []*RedisData
}

type ApiGetArchitectureInfoOfAPackageIn struct {
	PackageName string `annotation:"@empty=useAllPackages;"`
}
type ApiGetArchitectureInfoOfAPackageOut map[string]*ArchitechInfoOfAPackage

var APIGetArchitectureDetailOfPackages = api.Api(func(packInfo *ApiGetArchitectureInfoOfAPackageIn) (architects ApiGetArchitectureInfoOfAPackageOut, err error) {
	var (
		typeSpec   *ast.TypeSpec
		structType *ast.StructType
		ok         bool
		pkgs       map[string]*ast.Package
	)

	//get all package names
	if packInfo.PackageName == "" {
		packInfo.PackageName = "."
	}
	fset := token.NewFileSet()
	if pkgs, err = parser.ParseDir(fset, packInfo.PackageName, nil, parser.PackageClauseOnly); err != nil {
		return architects, err
	}

	architects = make(map[string]*ArchitechInfoOfAPackage)

	for _, pkg := range pkgs {
		architects[pkg.Name] = &ArchitechInfoOfAPackage{
			StructInfos: []*StructInfo{},
			TypeProtos:  []string{},
			FuncProtos:  []string{},
		}
		ast.Inspect(pkg, func(n ast.Node) bool {
			if typeSpec, ok = n.(*ast.TypeSpec); ok {
				if structType, ok = typeSpec.Type.(*ast.StructType); ok {
					var fields []string
					for _, field := range structType.Fields.List {
						var tag string
						var names []string
						if field.Tag != nil {
							tag = " `" + field.Tag.Value + "`"
						}
						// names joined with comma
						for _, name := range field.Names {
							names = append(names, name.Name)
						}

						fields = append(fields, strings.Join(names, ", ")+" "+field.Type.(*ast.Ident).Name+tag)
					}
					architects[pkg.Name].StructInfos = append(architects[pkg.Name].StructInfos, &StructInfo{
						StructName:            typeSpec.Name.Name,
						FieldsWithNameTypeTag: fields,
					})
				} else {
					// structType to string
					structTypeString := typeSpec.Type.(*ast.Ident).Name
					architects[pkg.Name].TypeProtos = append(architects[pkg.Name].TypeProtos, "type "+typeSpec.Name.Name+" "+structTypeString)
				}
			} else if funcDecl, ok := n.(*ast.FuncDecl); ok {
				var funcProto string
				funcProto = funcDecl.Name.Name + "("
				for _, param := range funcDecl.Type.Params.List {
					for _, name := range param.Names {
						funcProto += name.Name + " " + param.Type.(*ast.Ident).Name + ", "
					}
				}
				funcProto = funcProto[:len(funcProto)-2] + ") ("
				for _, result := range funcDecl.Type.Results.List {
					funcProto += result.Type.(*ast.Ident).Name + ", "
				}
				funcProto = funcProto[:len(funcProto)-2] + ")"
				architects[pkg.Name].FuncProtos = append(architects[pkg.Name].FuncProtos, funcProto)
			}
			return true

		})
	}
	fmt.Println(fmt.Sprintf("%+v", architects))
	return architects, nil
}).Func
