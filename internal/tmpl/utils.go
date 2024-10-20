package tmpl

import (
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

// 首字母小写驼峰
func UnderscoreToLowerCamelCase(s string) string {
	s = UnderscoreToUpperCamelCase(s)
	//  首字母小写
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}

// 下划线
func CamelCaseToUnderscore(s string) string {
	var output []rune
	for i, r := range s {
		if i == 0 {
			output = append(output, unicode.ToLower(r))
			continue
		}
		if unicode.IsUpper(r) && !unicode.IsUpper(rune(s[i-1])) {
			output = append(output, '_')
		}
		output = append(output, unicode.ToLower(r))
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
