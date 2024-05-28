package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Config 配置结构
type Config struct {
	ServiceCodes []CodeLabel `json:"service_codes"` // 服务级错误码
	ModuleCodes  []CodeLabel `json:"module_codes"`  // 模块级错误码
	I18n         []string    `json:"i18n"`          // 支持的语言列表
	FilePath     string      `json:"file_path"`     // 错误码文件路径
}

// CodeLabel 代码标签结构
type CodeLabel struct {
	Code  string `json:"code"`  // 代码
	Label string `json:"label"` // 标签
	Desc  string `json:"desc"`  // 描述
}

// 错误码常量到错误码的映射
var errorKeyCodeMap = make(map[string]int)

// 错误码到行号的映射
var errorCodeLine = make(map[int]int)

func main() {
	// 命令行工具的主入口
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Printf("配置文件加载错误: %v\n", err)
		return
	}

	serviceCode := selectFromConfig(config.ServiceCodes, "请选择服务级错误码:")
	moduleCode := selectFromConfig(config.ModuleCodes, "请选择模块级错误码:")

	fmt.Print("请输入错误码常量键: ")
	reader := bufio.NewReader(os.Stdin)
	errorCodeKey, _ := reader.ReadString('\n')
	errorCodeKey = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(errorCodeKey), " ", "_"))
	//fmt.Println("config==", config)
	if exists, _ := checkErrorCodeExists(config.FilePath, errorCodeKey); exists {
		fmt.Println("错误码键已存在，请输入不同的键.")
		return
	}

	messages := make(map[string]string)
	for _, lang := range config.I18n {
		fmt.Printf("为 %s 输入错误信息:\n", lang)
		msg, _ := reader.ReadString('\n')
		messages[lang] = strings.TrimSpace(msg)
	}

	updateProtoFile(config.FilePath, serviceCode, moduleCode, errorCodeKey, messages)
}

func loadConfig(path string) (*Config, error) {
	// 加载配置文件
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	return &config, err
}

func selectFromConfig(items []CodeLabel, prompt string) string {
	// 从配置中选择一个项目
	fmt.Println(prompt)
	for i, item := range items {
		fmt.Printf("%d: %s [%s]\n", i, item.Label, item.Desc)
	}

	fmt.Print("输入编号 (回车为默认): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return items[0].Code
	}

	index, err := strconv.Atoi(input)
	if err != nil || index < 0 || index >= len(items) {
		fmt.Println("输入无效, 选择默认.")
		return items[0].Code
	}
	return items[index].Code
}

func checkErrorCodeExists(filePath, errorCodeKey string) (bool, error) {
	// 检查错误码键是否存在
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return false, fmt.Errorf("创建目录错误: %v", err)
		}
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := initializeProtoFile(filePath); err != nil {
			return false, fmt.Errorf("初始化proto文件错误: %v", err)
		}
	}

	if err := parseProtoFile(filePath); err != nil {
		return false, fmt.Errorf("解析proto文件错误: %v", err)
	}

	_, exists := errorKeyCodeMap[errorCodeKey]
	return exists, nil
}

func initializeProtoFile(filePath string) error {
	fmt.Println("filePath==", filePath)
	// 确保文件所在目录存在
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录错误: %v", err)
		}
	}

	// 初始化proto文件的内容
	content := `syntax = "proto3";

package errcode;

option go_package = "err/errcode";
import "errors/errors.proto";

enum ErrorCode {
  UNKNOWN = 100000 [(errors.msg) = "未知错误", (errors.msg_english) = "unknown error"];

  // Business Logic Validation Errors (5xxxxx)
  INVALID_PARAMS = 200001 [(errors.msg) = "请求参数错误", (errors.msg_english) = "request parameter error"];
  NO_TOKEN = 200002 [(errors.msg) = "Token 不合法或者不存在", (errors.msg_english) = "token is invalid or does not exist"];
  
  LOGIN_FAILURE = 201001 [(errors.msg) = "用户名不存在或者密码错误!", (errors.msg_english) = "The username does not exist or the password is incorrect"];
  USER_BAN = 201002 [(errors.msg) = "用户被禁用", (errors.msg_english) = "user is banned"];
}
`
	// 尝试写入文件
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("写入文件错误: %v, 确保文件路径 '%s' 是正确的并且有写入权限", err, filePath)
	}

	return nil
}

func parseProtoFile(filePath string) error {
	// 解析proto文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	re := regexp.MustCompile(`^\s*(\w+)\s*=\s*(\d+)\s*\[`)
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++
		if matches := re.FindStringSubmatch(line); matches != nil {
			errorCodeKey := matches[1]
			errorCode := matches[2]
			codeInt, _ := strconv.Atoi(errorCode)
			errorKeyCodeMap[errorCodeKey] = codeInt
			errorCodeLine[codeInt] = lineNumber
		}
	}

	return scanner.Err()
}

func updateProtoFile(filePath, serviceCode, moduleCode, errorCodeKey string, messages map[string]string) {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("创建目录错误: %v\n", err)
			return
		}
	}

	// 检查文件是否存在，如果不存在则创建并初始化
	var content []byte
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := initializeProtoFile(filePath); err != nil {
			fmt.Printf("初始化proto文件错误: %v\n", err)
			return
		}
		content, _ = os.ReadFile(filePath) // 读取刚创建的文件
	} else {
		var err error
		content, err = os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("读取文件错误: %v\n", err)
			return
		}
	}

	// 更新或插入新的错误码定义
	updatedContent, err := insertOrUpdateErrorCode(string(content), errorCodeKey, serviceCode, moduleCode, messages)
	if err != nil {
		return
	}
	// 写回更新后的内容到 .proto 文件
	if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
		fmt.Printf("写入更新文件错误: %v\n", err)
		return
	}

	fmt.Println("更新errors.proto文件:", filePath)
}

func buildErrorCodeDefinition(serviceCode, moduleCode, errorCodeKey string, messages map[string]string) string {
	// 构建新的错误码定义
	newCode, _ := getNextErrorCode(serviceCode, moduleCode)
	fullErrorCode := serviceCode + moduleCode + newCode

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("\n  %s = %s [", errorCodeKey, fullErrorCode))

	keys := make([]string, 0, len(messages))
	for k := range messages {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, lang := range keys {
		msg := messages[lang]
		langKey := fmt.Sprintf("errors.msg_%s", strings.ToLower(lang))
		if lang == "default" {
			langKey = "errors.msg"
		}

		builder.WriteString(fmt.Sprintf("(%s) = \"%s\"", langKey, msg))
		if i < len(keys)-1 {
			builder.WriteString(", ")
		}
	}

	builder.WriteString("];")
	return builder.String()
}

func insertOrUpdateErrorCode(content, errorCodeKey, serviceCode, moduleCode string, messages map[string]string) (string, error) {

	// 检查新的错误码常量名是否已存在
	if _, ok := errorKeyCodeMap[errorCodeKey]; ok {
		return "", fmt.Errorf("错误码常量名 %s 已存在", errorCodeKey)
	}

	// 找到与新错误码前缀相同（前三位）且后三位最大值的行号
	prefix, _ := strconv.Atoi(serviceCode + moduleCode)
	prefixMax := prefix*1000 + 999
	maxCode := 0
	//fmt.Println("prefixMax===", prefixMax)
	//fmt.Println("errorKeyCodeMap", errorKeyCodeMap)
	for _, code := range errorKeyCodeMap {
		//fmt.Println("code===", code)
		if code > maxCode && code <= prefixMax {
			//fmt.Println("maxCode = code----", code)
			maxCode = code
		}
	}
	//fmt.Println("maxCode", maxCode)
	// 使用正则表达式匹配到指定的错误码行
	regStr := fmt.Sprintf(`(\s*) = %d.*;`, maxCode)
	regex1 := regexp.MustCompile(regStr)
	//fmt.Println("regex1", regex1)
	matches1 := regex1.FindStringSubmatch(content)
	//fmt.Println("matches1", matches1)
	if len(matches1) == 0 {
		return "", fmt.Errorf("未找到指定的错误码行")
	}
	// 构建新的错误码定义
	newErrorCodeDefinition := buildErrorCodeDefinition(serviceCode, moduleCode, errorCodeKey, messages)

	newLine := ""
	if maxCode/1000 != prefix {
		newLine = "\n"
	}

	// 在匹配到的错误码行后插入新的错误码信息
	content = strings.Replace(content, matches1[0], matches1[0]+newLine+newErrorCodeDefinition, 1)

	return content, nil
}

func getNextErrorCode(serviceCode, moduleCode string) (string, error) {
	// 生成新的错误码
	maxCode := 0
	prefix := serviceCode + moduleCode
	for code, _ := range errorCodeLine {
		codeStr := strconv.Itoa(code)
		if strings.HasPrefix(codeStr, prefix) {
			suffix, err := strconv.Atoi(codeStr[len(prefix):])
			if err != nil {
				continue
			}
			if suffix > maxCode {
				maxCode = suffix
			}
		}
	}
	newCode := maxCode + 1
	return fmt.Sprintf("%03d", newCode), nil
}
