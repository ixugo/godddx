package tmpl

import (
	"fmt"
	"strings"
	"unicode"
)

// 首字母大写驼峰
func UnderscoreToUpperCamelCase(s string) string {
	// 字符串替换, -1 表示不限制次数
	s = strings.Replace(s, "_", " ", -1)
	// 每个单词首字母大写
	s = strings.Title(s)
	return strings.Replace(s, " ", "", -1)
}

// 如果首字母是大写，则首字母大写驼峰
func IfUpperUnderscoreToUpperCamelCase(s string) string {
	if len(s) > 0 && unicode.IsLower(rune(s[0])) {
		return fmt.Sprintf("%c%s", rune(s[0]), UnderscoreToUpperCamelCase(s)[1:])
	}

	// 字符串替换, -1 表示不限制次数
	s = strings.Replace(s, "_", " ", -1)
	// 每个单词首字母大写
	s = strings.Title(s)
	return strings.Replace(s, " ", "", -1)
}

// 首字母小写驼峰
func UnderscoreToLowerCamelCase(s string) string {
	s = UnderscoreToUpperCamelCase(s)
	//  首字母小写
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}

// 下划线
func CamelCaseToUnderscore(s string) string {
	output := make([]rune, 0, len(s))
	var lastIsLower bool
	for _, r := range s {
		if lastIsLower && unicode.IsUpper(r) {
			output = append(output, '_')
		}
		output = append(output, unicode.ToLower(r))
		if !unicode.IsDigit(r) {
			lastIsLower = unicode.IsLower(r)
		}
	}
	return string(output)
}

func Plural(s string) string {
	s = CamelCaseToUnderscore(s)
	if strings.HasSuffix(s, "s") {
		return s
	}
	return s + "s"
}

func ToComment(s string) string {
	if s == "" {
		return ""
	}
	return "// " + s
}

func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// FirstLetter 首字母小写
func FirstLetter(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToLower(s[:1])
}
