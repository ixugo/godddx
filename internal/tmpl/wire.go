package tmpl

// 识别是否存在 internal/web/api/provider.go 文件，自动依赖注入

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FileExists 返回相对 baseDir 下是否存在 `internal/web/api/provider.go` 文件。
// baseDir 传入工作目录，例如 "."。
func FileExists(baseDir, filename string) bool {
	providerPath := filepath.Join(baseDir, "internal", "web", "api", filename)
	_, err := os.Stat(providerPath)
	return err == nil
}

// AppendProviderSetArg 在 `internal/web/api/provider.go` 内为 ProviderSet 的 wire.NewSet 调用追加参数（支持多个，自动去重）。
// - 优先定位 `ProviderSet = wire.NewSet(...)`，否则回退到文件内最后一个 `wire.NewSet(...)` 调用。
// - newArgExprs 为合法 Go 表达式（如："test.New"），若已存在则跳过，不重复插入。
// 返回值表示是否完成了追加操作，以及可能的错误。
func AppendProviderSetArg(baseDir string, newArgExprs ...string) (bool, error) {
	providerPath := filepath.Join(baseDir, "internal", "web", "api", "provider.go")
	srcBytes, err := os.ReadFile(providerPath)
	if err != nil {
		return false, err
	}
	fset := token.NewFileSet()
	fileAST, err := parser.ParseFile(fset, providerPath, srcBytes, parser.ParseComments)
	if err != nil {
		return false, err
	}

	// 追加基于文本插入，不再解析 newArgExpr

	var (
		targetCall     *ast.CallExpr
		lastWireNewSet *ast.CallExpr
		modified       bool
	)

	// 遍历 AST：优先寻找 ProviderSet 的 wire.NewSet 调用；同时记录文件中最后一个 wire.NewSet 调用
	ast.Inspect(fileAST, func(n ast.Node) bool {
		// 记录所有 wire.NewSet 调用，保留最后一个
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel != nil && sel.Sel.Name == "NewSet" {
					if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "wire" {
						lastWireNewSet = call
					}
				}
			}
		}

		// 查找 var ProviderSet = wire.NewSet(...)
		gen, ok := n.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			return true
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			// 寻找名为 ProviderSet 的声明
			hasProviderSetName := false
			for _, name := range vs.Names {
				if name != nil && name.Name == "ProviderSet" {
					hasProviderSetName = true
					break
				}
			}
			if !hasProviderSetName || len(vs.Values) == 0 {
				continue
			}
			// 检查是否为 wire.NewSet 调用
			if call, ok := vs.Values[0].(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if sel.Sel != nil && sel.Sel.Name == "NewSet" {
						if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "wire" {
							targetCall = call
							return false // 已找到目标，停止深入当前分支
						}
					}
				}
			}
		}
		return true
	})

	// 选择目标调用顺序：优先 ProviderSet，其次文件中最后一个 wire.NewSet
	callToEdit := targetCall
	if callToEdit == nil {
		callToEdit = lastWireNewSet
	}
	if callToEdit == nil {
		// 文件中不存在 wire.NewSet 调用
		return false, nil
	}

	// 构建已存在参数集合，基于源码切片进行比较，避免误判
	existing := make(map[string]struct{}, len(callToEdit.Args))
	for _, a := range callToEdit.Args {
		start := fset.Position(a.Pos()).Offset
		end := fset.Position(a.End()).Offset
		if start >= 0 && end <= len(srcBytes) && start < end {
			s := string(bytes.TrimSpace(srcBytes[start:end]))
			// 去掉末尾逗号
			if len(s) > 0 && s[len(s)-1] == ',' {
				s = s[:len(s)-1]
			}
			existing[s] = struct{}{}
		}
	}

	// 去重输入并过滤已存在
	uniqToInsert := make([]string, 0, len(newArgExprs))
	seen := make(map[string]struct{}, len(newArgExprs))
	for _, arg := range newArgExprs {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}
		if _, ok := seen[arg]; ok {
			continue
		}
		seen[arg] = struct{}{}
		if _, ok := existing[arg]; ok {
			continue
		}
		uniqToInsert = append(uniqToInsert, arg)
	}

	if len(uniqToInsert) == 0 {
		return false, nil
	}

	// 将缺失的参数以换行的形式一次性插入到目标调用的右括号之前
	endOffset := fset.Position(callToEdit.End()).Offset
	if endOffset == 0 || endOffset > len(srcBytes) {
		return false, errors.New("invalid call position")
	}
	insertAt := endOffset - 1 // 右括号位置
	var insertion bytes.Buffer
	insertion.WriteString("\t")
	for _, arg := range uniqToInsert {
		insertion.WriteString(arg)
		insertion.WriteString(", ")
	}
	insertion.WriteString("\n")

	var out bytes.Buffer
	out.Write(srcBytes[:insertAt])
	out.Write(insertion.Bytes())
	out.Write(srcBytes[insertAt:])
	if err := os.WriteFile(providerPath, out.Bytes(), 0o600); err != nil {
		return false, err
	}
	modified = true

	if !modified {
		return false, nil
	}
	return true, nil
}

// AppendLineToSetupRouter 在 `internal/web/api/api.go` 中定位 `setupRouter` 函数，并在函数体结束前插入一行代码。
// newLine 必须为合法 Go 语句（无需结尾分号），例如：`registerTmpl(r, NewTmplAPIFromDB(uc.DB))`
func AppendLineToSetupRouter(baseDir string, funcName, newLine string) (bool, error) {
	apiPath := filepath.Join(baseDir, "internal", "web", "api", "api.go")
	if _, err := os.Stat(apiPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	srcBytes, err := os.ReadFile(apiPath)
	if err != nil {
		return false, err
	}

	fset := token.NewFileSet()
	fileAST, err := parser.ParseFile(fset, apiPath, srcBytes, 0)
	if err != nil {
		return false, err
	}

	// 找到 setupRouter 函数
	var (
		bodyRBrace token.Pos
		setupFn    *ast.FuncDecl
		found      bool
	)
	for _, decl := range fileAST.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Name != nil && fn.Name.Name == "setupRouter" && fn.Body != nil {
				setupFn = fn
				bodyRBrace = fn.Body.Rbrace
				found = true
				break
			}
		}
	}
	if !found {
		return false, nil
	}

	// 检查 setupRouter 函数体内是否已调用 funcName（支持 Ident 或 SelectorExpr.Sel）
	called := false
	ast.Inspect(setupFn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			if fun.Name == funcName {
				called = true
				return false
			}
		case *ast.SelectorExpr:
			if fun.Sel != nil && fun.Sel.Name == funcName {
				called = true
				return false
			}
		}
		return true
	})
	if called {
		return false, nil
	}

	// 在右大括号前插入一行
	rbraceOffset := fset.Position(bodyRBrace).Offset
	if rbraceOffset == 0 || rbraceOffset > len(srcBytes) {
		return false, errors.New("invalid setupRouter body position")
	}

	// 计算缩进：取右大括号所在行的前导空白作为缩进基准
	lineStart := rbraceOffset - 1
	for lineStart > 0 && srcBytes[lineStart] != '\n' {
		lineStart--
	}

	insertion := "\t// TODO: 待补充中间件\n\t" + newLine + "\n"

	var out bytes.Buffer
	out.Write(srcBytes[:rbraceOffset])
	out.Write([]byte(insertion))
	out.Write(srcBytes[rbraceOffset:])
	if err := os.WriteFile(apiPath, out.Bytes(), 0o600); err != nil {
		return false, err
	}
	return true, nil
}

// AppendUsecaseField 在 `internal/web/api/provider.go` 中找到 `type Usecase struct { ... }`，
// 并在右大括号之前插入一行字段定义。fieldDecl 示例：`NewProp string` 或 `Svc service.API`。
func AppendUsecaseField(baseDir string, fieldDecl string) (bool, error) {
	providerPath := filepath.Join(baseDir, "internal", "web", "api", "provider.go")
	srcBytes, err := os.ReadFile(providerPath)
	if err != nil {
		return false, err
	}

	fset := token.NewFileSet()
	fileAST, err := parser.ParseFile(fset, providerPath, srcBytes, 0)
	if err != nil {
		return false, err
	}

	var (
		rbrace token.Pos
		found  bool
	)
	// 遍历找到 Usecase 结构体
	for _, decl := range fileAST.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name == nil || ts.Name.Name != "Usecase" {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok || st.Fields == nil || !st.Fields.Closing.IsValid() {
				continue
			}

			// 去重：若字段已存在（同名），则跳过
			if st.Fields != nil {
				for _, f := range st.Fields.List {
					if len(f.Names) > 0 {
						existName := f.Names[0].Name
						// 从 fieldDecl 提取字段名（第一个 token，遇到空格或制表符截止）
						fd := strings.TrimSpace(fieldDecl)
						for i := 0; i < len(fd); i++ {
							if fd[i] == ' ' || fd[i] == '\t' {
								fd = fd[:i]
								break
							}
						}
						if fd == existName {
							return false, nil
						}
					}
				}
			}

			rbrace = st.Fields.Closing
			found = true
			break
		}
		if found {
			break
		}
	}

	if !found {
		return false, nil
	}

	// 在结构体右大括号前插入一行：\n\tfieldDecl
	offset := fset.Position(rbrace).Offset
	if offset == 0 || offset > len(srcBytes) {
		return false, errors.New("invalid struct position")
	}
	insertion := []byte("\n\t" + fieldDecl + "\n")
	var out bytes.Buffer
	out.Write(srcBytes[:offset])
	out.Write(insertion)
	out.Write(srcBytes[offset:])
	if err := os.WriteFile(providerPath, out.Bytes(), 0o600); err != nil {
		return false, err
	}

	return true, nil
}

func MakeWire() error {
	return CommandContext("wire", "./...")
}

func CommandContext(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%s %s", args[0], err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s %s", args[0], err)
	}
	return nil
}
