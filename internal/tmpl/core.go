package tmpl

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed api.go.tmpl core.go.tmpl db.go.tmpl model.go.tmpl db.engine.go.tmpl core.engine.go.tmpl model.engine.go.tmpl param.engine.go.tmpl
var files embed.FS

type Data struct {
	Name        []Name
	PackageName string // 包名，其实是首字母小写的 domain
}

type Name struct {
	TableName   string
	PackageName string
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

	// api
	{

		tp, err := generateModelCode(domain)
		if err != nil {
			return err
		}
		apiFile := bytes.NewBuffer(nil)
		out[fmt.Sprintf("internal/web/api/%s.go", domain.PackageName)] = apiFile

		tpl := template.Must(template.New("abc").Funcs(
			template.FuncMap{
				"ToUpperCamelCase":                  UnderscoreToUpperCamelCase, // 首字母大写驼峰
				"ToLowerCamelCase":                  UnderscoreToLowerCamelCase, // 首字母小写驼峰
				"ToUnderscore":                      CamelCaseToUnderscore,      // 蛇形
				"Plural":                            Plural,
				"ToComment":                         ToComment,
				"IfUpperUnderscoreToUpperCamelCase": IfUpperUnderscoreToUpperCamelCase,
			},
		).ParseFS(files, "api.go.tmpl", "db.go.tmpl"))

		if err := tpl.ExecuteTemplate(apiFile, "api.go.tmpl", tp); err != nil {
			panic(err)
		}

		// 写到硬盘
		for k, v := range out {
			_ = os.MkdirAll(filepath.Dir(k), os.ModePerm)
			os.WriteFile(k, v.Bytes(), os.ModePerm)
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

	tpl := template.Must(template.New("abc").Funcs(
		template.FuncMap{
			"ToUpperCamelCase":                  UnderscoreToUpperCamelCase, // 首字母大写驼峰
			"ToLowerCamelCase":                  UnderscoreToLowerCamelCase, // 首字母小写驼峰
			"ToUnderscore":                      CamelCaseToUnderscore,      // 蛇形
			"Plural":                            Plural,
			"ToComment":                         ToComment,
			"IfUpperUnderscoreToUpperCamelCase": IfUpperUnderscoreToUpperCamelCase,
		},
	).ParseFS(files, "model.go.tmpl", "model.engine.go.tmpl"))

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

	tpl := template.Must(template.New("abc").Funcs(
		template.FuncMap{
			"ToUpperCamelCase":                  UnderscoreToUpperCamelCase, // 首字母大写驼峰
			"ToLowerCamelCase":                  UnderscoreToLowerCamelCase, // 首字母小写驼峰
			"ToUnderscore":                      CamelCaseToUnderscore,      // 蛇形
			"Plural":                            Plural,
			"ToComment":                         ToComment,
			"IfUpperUnderscoreToUpperCamelCase": IfUpperUnderscoreToUpperCamelCase,
		},
	).ParseFS(files, "core.go.tmpl", "core.engine.go.tmpl", "param.engine.go.tmpl"))

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

	tpl := template.Must(template.New("abc").Funcs(
		template.FuncMap{
			"ToUpperCamelCase":                  UnderscoreToUpperCamelCase, // 首字母大写驼峰
			"ToLowerCamelCase":                  UnderscoreToLowerCamelCase, // 首字母小写驼峰
			"ToUnderscore":                      CamelCaseToUnderscore,      // 蛇形
			"Plural":                            Plural,
			"ToComment":                         ToComment,
			"IfUpperUnderscoreToUpperCamelCase": IfUpperUnderscoreToUpperCamelCase,
		},
	).ParseFS(files, "db.engine.go.tmpl", "db.go.tmpl"))

	if err := tpl.ExecuteTemplate(buf, "db.go.tmpl", tp); err != nil {
		panic(err)
	}
	bufMap[fmt.Sprintf("internal/core/%s/store/%sdb/db.go", out.PackageName, out.PackageName)] = buf

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
	}

	return nil
}
