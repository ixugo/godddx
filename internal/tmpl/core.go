package tmpl

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed *.go.tmpl
var files embed.FS

type Data struct {
	Name        []Name
	PackageName string // 包名，其实是首字母小写的 domain
}

type Name struct {
	TableName   string
	PackageName string
}

var funcMap = template.FuncMap{
	"ToUpperCamelCase":                  UnderscoreToUpperCamelCase, // 首字母大写驼峰
	"ToLowerCamelCase":                  UnderscoreToLowerCamelCase, // 首字母小写驼峰
	"ToUnderscore":                      CamelCaseToUnderscore,      // 蛇形
	"Plural":                            Plural,
	"ToComment":                         ToComment,
	"IfUpperUnderscoreToUpperCamelCase": IfUpperUnderscoreToUpperCamelCase,
	"ToUpper":                           ToUpper,
	"FirstLetter":                       FirstLetter,
}

func Start(path, module string) error {
	domain, err := ParseFile(path)
	if err != nil {
		return err
	}
	domain.ModuleName = module
	// 虚拟目录
	out := make(map[string]*bytes.Buffer)

	// core/model.go
	if err := handlerDomainModel(domain, out); err != nil {
		return err
	}

	// core/core.go
	if err := handlerDomainCore(domain, out); err != nil {
		return err
	}
	// core/store/userdb/db.go
	if err := handlerDomainDB(domain, out); err != nil {
		return err
	}
	// core/store/usercache/cache.go
	if err := handlerDomainCache(domain, out); err != nil {
		return err
	}

	// api
	{

		tp, err := generateModelCode(domain)
		if err != nil {
			return err
		}
		apiFile := bytes.NewBuffer(nil)
		out[fmt.Sprintf("internal/web/api/%s.go", domain.PackageName)] = apiFile

		tpl := template.Must(template.New("abc").Funcs(funcMap).
			ParseFS(files, "api.go.tmpl", "db.go.tmpl"))

		if err := tpl.ExecuteTemplate(apiFile, "api.go.tmpl", tp); err != nil {
			panic(err)
		}

		// 写到硬盘
		for k, v := range out {
			_ = os.MkdirAll(filepath.Dir(k), os.ModePerm)
			if err := os.WriteFile(k, v.Bytes(), os.ModePerm); err != nil {
				fmt.Println("⚠️ WriteFile err:", err)
			}
		}
	}

	// 填充 provider.go 依赖注入
	if FileExists("", "provider.go") {
		const uniqueidName = "NewUniqueID"
		if domain.ExistsUniqueID {
			if _, err := AppendProviderSetArg("", uniqueidName); err != nil {
				return fmt.Errorf("缺少 NewUniqueID, 请手动更新 provider.go 依赖注入, %w", err)
			}
		}

		apiName := fmt.Sprintf("New%sAPI", UnderscoreToUpperCamelCase(domain.PackageName))
		coreName := fmt.Sprintf("New%sCore", UnderscoreToUpperCamelCase(domain.PackageName))
		if _, err := AppendProviderSetArg("", coreName, apiName); err != nil {
			return fmt.Errorf("请手动更新 provider.go 依赖注入, %w", err)
		}

		fieldName := fmt.Sprintf("%sAPI", UnderscoreToUpperCamelCase(domain.PackageName))
		if _, err := AppendUsecaseField("", fmt.Sprintf("%s %s", fieldName, fieldName)); err != nil {
			return fmt.Errorf("请手动更新 provider.go 依赖注入, %w", err)
		}
		if err := MakeWire(); err != nil {
			fmt.Println("请手动执行 make wire, err:", err)
		}

		// 填充 api 路由
		if FileExists("", "api.go") {
			funcName := fmt.Sprintf("Register%s", UnderscoreToUpperCamelCase(domain.PackageName))
			line := fmt.Sprintf("%s(r, uc.%s)", funcName, fieldName)
			if _, err := AppendLineToSetupRouter("", funcName, line); err != nil {
				return fmt.Errorf("请手动更新 api.go 路由, %w", err)
			}
		}
	}

	return nil
}

func handlerDomainModel(out *Domain, bufMap map[string]*bytes.Buffer) error {
	tp, err := generateModelCode(out)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)

	tpl := template.Must(template.New("abc").Funcs(funcMap).ParseFS(files, "model.go.tmpl", "model.engine.go.tmpl"))

	if err := tpl.ExecuteTemplate(buf, "model.go.tmpl", tp); err != nil {
		panic(err)
	}
	bufMap[fmt.Sprintf("internal/core/%s/model.go", out.PackageName)] = buf

	for _, v := range tp.Models {
		if v.IsNotDB {
			continue
		}

		v.PackageName = out.PackageName
		buf := bytes.NewBuffer(nil)
		if err := tpl.ExecuteTemplate(buf, "model.engine.go.tmpl", v); err != nil {
			panic(err)
		}
		bufMap[fmt.Sprintf("internal/core/%s/%s.model.go", out.PackageName, CamelCaseToUnderscore(v.Name))] = buf
	}

	return nil
}

func handlerDomainCore(out *Domain, bufMap map[string]*bytes.Buffer) error {
	tp, err := generateModelCode(out)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)

	tpl := template.Must(template.New("abc").Funcs(funcMap).ParseFS(files, "core.go.tmpl", "core.engine.go.tmpl", "param.engine.go.tmpl"))

	if err := tpl.ExecuteTemplate(buf, "core.go.tmpl", tp); err != nil {
		panic(err)
	}

	bufMap[fmt.Sprintf("internal/core/%s/core.go", out.PackageName)] = buf

	for _, v := range tp.Models {
		if v.IsNotDB {
			continue
		}

		v.PackageName = out.PackageName
		buf := bytes.NewBuffer(nil)
		if err := tpl.ExecuteTemplate(buf, "core.engine.go.tmpl", v); err != nil {
			panic(err)
		}
		bufMap[fmt.Sprintf("internal/core/%s/%s.go", out.PackageName, CamelCaseToUnderscore(v.Name))] = buf
	}

	for _, v := range tp.Models {
		if v.IsNotDB {
			continue
		}

		v.PackageName = out.PackageName
		buf := bytes.NewBuffer(nil)
		if err := tpl.ExecuteTemplate(buf, "param.engine.go.tmpl", v); err != nil {
			panic(err)
		}
		bufMap[fmt.Sprintf("internal/core/%s/%s.param.go", out.PackageName, CamelCaseToUnderscore(v.Name))] = buf
	}

	return nil
}

func handlerDomainDB(out *Domain, bufMap map[string]*bytes.Buffer) error {
	tp, err := generateModelCode(out)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)

	tpl := template.Must(template.New("abc").Funcs(funcMap).ParseFS(files, "db.engine.go.tmpl", "db.go.tmpl", "db_test.go.tmpl", "db.engine_test.go.tmpl"))

	if err := tpl.ExecuteTemplate(buf, "db.go.tmpl", tp); err != nil {
		panic(err)
	}
	bufMap[fmt.Sprintf("internal/core/%s/store/%sdb/db.go", out.PackageName, out.PackageName)] = buf

	{
		// 移除测试
		// dbtestBuf := bytes.NewBuffer(nil)
		// if err := tpl.ExecuteTemplate(dbtestBuf, "db_test.go.tmpl", tp); err != nil {
		// 	panic(err)
		// }
		// bufMap[fmt.Sprintf("internal/core/%s/store/%sdb/db_test.go", out.PackageName, out.PackageName)] = dbtestBuf
	}

	for _, v := range tp.Models {
		if v.IsNotDB {
			continue
		}

		v.PackageName = out.PackageName
		buf := bytes.NewBuffer(nil)
		if err := tpl.ExecuteTemplate(buf, "db.engine.go.tmpl", v); err != nil {
			panic(err)
		}
		bufMap[fmt.Sprintf("internal/core/%s/store/%sdb/%s.go", out.PackageName, out.PackageName, CamelCaseToUnderscore(v.Name))] = buf

		{
			// 移除测试
			// dbengineBuf := bytes.NewBuffer(nil)
			// if err := tpl.ExecuteTemplate(dbengineBuf, "db.engine_test.go.tmpl", v); err != nil {
			// 	panic(err)
			// }
			// bufMap[fmt.Sprintf("internal/core/%s/store/%sdb/%s_test.go", out.PackageName, out.PackageName, CamelCaseToUnderscore(v.Name))] = dbengineBuf
		}
	}

	return nil
}

func handlerDomainCache(out *Domain, bufMap map[string]*bytes.Buffer) error {
	tp, err := generateModelCode(out)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)

	tpl := template.Must(template.New("abc").Funcs(funcMap).ParseFS(files, "cache.go.tmpl", "cache.engine.go.tmpl"))

	if err := tpl.ExecuteTemplate(buf, "cache.go.tmpl", tp); err != nil {
		panic(err)
	}
	bufMap[fmt.Sprintf("internal/core/%s/store/%scache/cache.go", out.PackageName, out.PackageName)] = buf

	for _, v := range tp.Models {
		if v.IsNotDB {
			continue
		}

		v.PackageName = out.PackageName
		buf := bytes.NewBuffer(nil)
		if err := tpl.ExecuteTemplate(buf, "cache.engine.go.tmpl", v); err != nil {
			panic(err)
		}
		bufMap[fmt.Sprintf("internal/core/%s/store/%scache/%s.go", out.PackageName, out.PackageName, CamelCaseToUnderscore(v.Name))] = buf
	}

	return nil
}
