package tmpl

import (
	"bytes"
	"fmt"
	"log"
	"testing"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TestParseFile(t *testing.T) {
	out, err := ParseFile("internal/tmpl/aa.go")
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%+v\n", out)
	tp, err := generateModelCode(out)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%+v\n", tp)

	buf := bytes.NewBuffer(nil)

	tpl := template.Must(template.New("abc").Funcs(
		template.FuncMap{
			"ToUpperCamelCase": UnderscoreToUpperCamelCase, // 首字母大写驼峰
			"ToLowerCamelCase": UnderscoreToLowerCamelCase, // 首字母小写驼峰
			"ToUnderscore":     CamelCaseToUnderscore,      // 蛇形
		},
	).ParseFS(files, "model.tmpl"))

	if err := tpl.ExecuteTemplate(buf, "model.tmpl", tp); err != nil {
		panic(err)
	}

	formattedCode, err := formatGoCode(buf.String())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(formattedCode)
}

func TestCamelCaseToUnderscore(t *testing.T) {
	args := []string{"ParentID", "ParentIDiDDC", "P2P", "P2Ps"}
	for _, arg := range args {

		s := CamelCaseToUnderscore(arg)
		fmt.Println(s)

		fmt.Println(cases.Upper(language.English).String(arg))

	}
}
