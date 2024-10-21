package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/ixugo/gowebx/internal/tmpl"
)

// CheckAndExtractModuleName 判断当前文件夹下是否存在 go.mod 并提取 module 名称
func CheckAndExtractModuleName() string {
	// 检查 go.mod 文件是否存在
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return ""
	}

	// 打开 go.mod 文件
	file, err := os.Open("go.mod")
	if err != nil {
		return ""
	}
	defer file.Close()

	// 逐行读取文件内容
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 判断行是否以 "module " 开头
		if strings.HasPrefix(line, "module ") {
			// 提取 "module " 后面的内容并去掉两边的空格
			moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			return moduleName
		}
	}
	return ""
}

var (
	file   = flag.String("f", "", "领域模型文件，多个用逗号分隔")
	module = flag.String("m", "", "模块名")
)

func main() {
	flag.Parse()

	moduleName := *module
	if moduleName == "" {
		moduleName = CheckAndExtractModuleName() //  `github.com/ixugo/gowebx`
	}
	if moduleName == "" {
		fmt.Println("未指定模块名称，请在 go.mod 同目录下执行，或者使用 -m 来指定模块名称")
		return
	}
	files := strings.Split(*file, ",")
	if len(files) == 0 {
		fmt.Println("未指定领域模型文件")
		return
	}
	for _, file := range files {
		if file == "" {
			continue
		}
		if err := tmpl.Start(file, moduleName); err != nil {
			slog.Error(err.Error())
		}
	}

	if err := CommandContext("goimports", "-w", "."); err != nil {
		// fmt.Println(err)
	}

	if err := CommandContext("gofumpt", "-l", "-w", "."); err != nil {
		// fmt.Println(err)
	}
}

func CommandContext(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}
