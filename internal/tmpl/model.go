package tmpl

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
)

// 模板类
type ModelTmpl struct {
	PackageName string
	ModuleName  string
	Models      []Table
}

type Table struct {
	PackageName string
	Name        string
	Lines       []Line
	IsNotDB     bool
}

type Line struct {
	Name    string
	Type    string
	Tag     string
	Comment string
}

// 提取 go 文件的结构体
type Domain struct {
	ModuleName  string
	PackageName string
	Models      []*Models
}

type Models struct {
	Name    string
	Ident   []Attribute
	IsNotDB bool // 顶层结构体默认 true，别人的属性，默认 false
}

type Attribute struct {
	Name    string
	Type    ast.Expr
	Model   *Models
	Comment string
}

func ParseFile(path string) (*Domain, error) {
	fileSet := token.NewFileSet()

	node, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return nil, err
	}

	var out Domain
	out.PackageName = node.Name.Name

	ast.Inspect(node, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					var model Models
					model.Name = typeSpec.Name.Name

					// 遍历结构体的字段
					for _, field := range structType.Fields.List {
						// 每个字段可能有多个名称
						for _, name := range field.Names {
							comment := strings.TrimSpace(field.Comment.Text())
							model.Ident = append(model.Ident, Attribute{
								Name:    name.Name,
								Type:    field.Type,
								Comment: comment,
							})
						}
					}
					out.Models = append(out.Models, &model)
				}
			}
		}
		return true
	})
	return &out, nil
}

func getStructName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name // 返回标识符名称
	case *ast.SelectorExpr:
		return e.Sel.Name // 返回选择器的名称
	}
	return ""
}

func generateModelCode(domain *Domain) (*ModelTmpl, error) {
	// 提取所有的结构体名
	// 判断该结构体名是否是别人的属性类型
	structs := make(map[string]*Models)
	for _, model := range domain.Models {
		structs[model.Name] = model
	}

	for _, model := range domain.Models {
		for _, filed := range model.Ident {
			_, ok := filed.Type.(*ast.StructType)
			if ok {
				continue
			}
			aname := getStructName(filed.Type)
			if f, ok := structs[aname]; ok {
				f.IsNotDB = true
			}
		}
	}

	otmpl := ModelTmpl{
		PackageName: domain.PackageName,
		ModuleName:  domain.ModuleName,
	}

	for _, model := range domain.Models {
		lines := make([]Line, 0, 8)
		for _, ident := range model.Ident {

			var tag strings.Builder
			defaultValue := generateTagGormDefaultValue(ident.Type)
			if defaultValue != "" {
				tag.WriteString(";default:" + defaultValue)
			}
			if defaultValue == "'{}'" {
				tag.WriteString(";type:jsonb")
			}
			if ident.Comment != "" {
				tag.WriteString(";comment:" + ident.Comment)
			}
			line := Line{
				Name:    ident.Name,
				Type:    fieldTypeToString(ident.Type),
				Tag:     fmt.Sprintf(`gorm:"column:%s;notNull%s"`, CamelCaseToUnderscore(ident.Name), tag.String()),
				Comment: ident.Comment,
			}
			if ident.Name == "ID" {
				line.Tag = `gorm:"primaryKey"`
			}

			lines = append(lines, line)
		}
		otmpl.Models = append(otmpl.Models, Table{
			Name:    model.Name,
			Lines:   lines,
			IsNotDB: model.IsNotDB,
		})
	}

	return &otmpl, nil
}

func generateTagGormDefaultValue(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		// 检查标识符是否是特定的类型
		switch e.Name {
		case "int", "int8", "int16", "int32", "int64", "float32", "float64":
			return "0"
		case "string":
			return "''"
		case "bool":
			return "FALSE"
		case "time":
			return "CURRENT_TIMESTAMP"
		default:
			return "'{}'"
		}
		// 底下都走不到
	case *ast.BasicLit:
		// 判断基础字面量类型
		switch e.Kind {
		case token.INT:
			return "0"
		case token.STRING:
			return ""
			// case token.bo:
			// return "FALSE"
		}
	case *ast.StructType:
		return "'{}'"
	case *ast.SelectorExpr:
		// 处理选择器（例如，time.Time）
		if e.X != nil && e.Sel != nil {
			return "CURRENT_TIMESTAMP" // 假设时间类型
		}
	}

	return "" // 对于其他未处理的类型返回空字符串
}

// 辅助函数，用于将字段类型转换为字符串
func fieldTypeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		// 处理带有包名的类型，如 time.Time
		return fmt.Sprintf("%s.%s", fieldTypeToString(t.X), t.Sel.Name)
	case *ast.ArrayType:
		// 处理数组类型
		return "[]" + fieldTypeToString(t.Elt)
	case *ast.StarExpr:
		// 处理指针类型
		return "*" + fieldTypeToString(t.X)
	case *ast.MapType:
		// 处理 map 类型
		return fmt.Sprintf("map[%s]%s", fieldTypeToString(t.Key), fieldTypeToString(t.Value))
	case *ast.FuncType:
		// 处理函数类型
		return "func(...)"
	default:
		return "unknown"
	}
}

func formatGoCode(sourceCode string) (string, error) {
	// 创建一个新的文件集
	fset := token.NewFileSet()

	// 解析源代码
	node, err := parser.ParseFile(fset, "", sourceCode, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("error parsing code: %w", err)
	}

	// 使用一个缓冲区来保存格式化后的代码
	var buf bytes.Buffer

	// 格式化并写入缓冲区
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return "", fmt.Errorf("error formatting code: %w", err)
	}

	return buf.String(), nil
}
