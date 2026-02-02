package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ixugo/godddx/internal/tmpl"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
	file      = flag.String("f", "", "领域模型文件，多个用逗号分隔")
	module    = flag.String("m", "", "模块名")
	version   = flag.Bool("v", false, "版本号")
	mcpServer = flag.Bool("mcp", false, "启动 MCP 服务")
	_         = flag.String("i", "", "1. 结构体想使用字符串 ID, 可考虑 ID uniqueid.Core, 自动生成全局唯一 ID")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println("github.com/ixugo/godddx v1.5.3")
		return
	}

	// 启动 MCP 服务模式
	if *mcpServer {
		runMCPServer()
		return
	}

	// 原有的命令行模式
	runCLI()
}

const description = `六边形架构 DDD 代码生成工具。

## 何时使用
- 创建新领域/实体的 CRUD 代码
- 新增数据库表及其完整业务层

## 生成内容
自动生成完整的分层代码：
- Core 层: core.go, <entity>.go, <entity>.model.go, <entity>.param.go
- Store 层: store/<domain>db/
- Cache 层: store/<domain>cache/
- API 层: internal/web/api/<domain>.go

## 使用方式
1. content: Go 模型定义，必须包含 Package 声明和 type 结构体定义，结构体必须包含 ID、CreatedAt、UpdatedAt 字段
2. module: Go 模块名，选填, 如 github.com/yourname/project
3. output_dir: 项目根目录（该目录下必须包含 go.mod 文件）

## 结构体示例
package user

import "github.com/ixugo/goddd/pkg/orm"

type User struct {
    ID        int      // 主键
    CreatedAt orm.Time
    UpdatedAt orm.Time
    Name      string   // 字段注释会转为 GORM comment
}

注: 随机字符串 ID 可用 uniqueid.Core 类型`

// runMCPServer 启动 MCP 服务，让 LLM 可以调用代码生成功能
func runMCPServer() {
	s := server.NewMCPServer(
		"godddx",
		"1.5.3",
		server.WithToolCapabilities(false),
	)

	// 添加代码生成工具
	tool := mcp.NewTool("generate_ddd_code",
		mcp.WithDescription(description),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Go 文件内容，必须包含 package 声明和 type 结构体定义"),
		),
		mcp.WithString("module",
			mcp.Required(),
			mcp.Description("选填，Go 模块名称，例如 github.com/yourname/project"),
		),
		mcp.WithString("output_dir",
			mcp.Required(),
			mcp.Description("项目根目录路径，即包含 go.mod 的目录"),
		),
	)

	s.AddTool(tool, generateCodeHandler)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "MCP Server error: %v\n", err)
		os.Exit(1)
	}
}

// generateCodeHandler 处理代码生成请求
func generateCodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	content, err := request.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError("参数错误: content 是必填项"), nil
	}

	moduleName, err := request.RequireString("module")
	if err != nil {
		return mcp.NewToolResultError("参数错误: module 是必填项"), nil
	}

	outputDir, err := request.RequireString("output_dir")
	if err != nil {
		return mcp.NewToolResultError("参数错误: output_dir 是必填项"), nil
	}

	// 切换到目标目录执行代码生成
	oldDir, err := os.Getwd()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取当前目录失败: %v", err)), nil
	}

	if err := os.Chdir(outputDir); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("切换到目标目录失败: %v", err)), nil
	}
	defer func() { _ = os.Chdir(oldDir) }()

	// 调用代码生成核心逻辑
	generatedFiles, err := tmpl.StartFromContent(content, moduleName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("代码生成失败: %v", err)), nil
	}

	// 对生成的代码执行格式化
	if err := tmpl.CommandContext("goimports", "-w", "."); err != nil {
		// goimports 失败不阻断，仅记录警告
		fmt.Fprintf(os.Stderr, "⚠️ goimports 执行失败: %v\n", err)
	}
	if err := tmpl.CommandContext("gofumpt", "-l", "-w", "."); err != nil {
		// gofumpt 失败不阻断，仅记录警告
		fmt.Fprintf(os.Stderr, "⚠️ gofumpt 执行失败: %v\n", err)
	}

	// 构建成功消息
	var result strings.Builder
	result.WriteString("代码生成成功！\n\n生成的文件列表:\n")
	for _, f := range generatedFiles {
		result.WriteString(fmt.Sprintf("  - %s\n", f))
	}

	return mcp.NewToolResultText(result.String()), nil
}

// runCLI 运行原有的命令行模式
func runCLI() {
	moduleName := *module
	if moduleName == "" {
		moduleName = CheckAndExtractModuleName()
	}
	if moduleName == "" {
		fmt.Println("未指定模块名称，请在 go.mod 同目录下执行，或者使用 -m 来指定模块名称")
		return
	}
	files := strings.Split(*file, ",")
	if len(files) == 0 {
		fmt.Println("⚠️ 未指定领域模型文件")
		return
	}
	for _, file := range files {
		if file == "" {
			continue
		}
		if err := tmpl.Start(file, moduleName); err != nil {
			fmt.Println("⚠️  err:", err)
		}
	}

	if err := tmpl.CommandContext("goimports", "-w", "."); err != nil {
		fmt.Println("⚠️  err:", err)
	}

	if err := tmpl.CommandContext("gofumpt", "-l", "-w", "."); err != nil {
		fmt.Println("⚠️  err:", err)
	}
}
