package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	InputPath  string
	OutputPath string
	NamePrefix string
	NameFormat string
	StartNum   int
}

const defaultIniPath = "./convert_okx.ini"
const defaultInputPath = "./"
const defaultOutputPath = "./rename"
const defaultNamePrefix = "NFT"
const defaultNameFormat = "{0} #{1}"
const defaultStartNum = 1

func main() {
	var iniPath, inputPath, outputPath, namePrefix, nameFormat string
	var showVer bool
	var startNum int
	flag.StringVar(&iniPath, "c", defaultIniPath, "ini配置文件路径")
	flag.StringVar(&inputPath, "i", defaultInputPath, "输入文件或目录路径")
	flag.StringVar(&outputPath, "o", defaultOutputPath, "输出目录路径")
	flag.StringVar(&namePrefix, "p", defaultNamePrefix, "名称前缀")
	flag.StringVar(&nameFormat, "f", defaultNameFormat, "名称格式, {0}为前缀占位符, {1}为编号数字占位符")
	flag.IntVar(&startNum, "n", defaultStartNum, "起始编号数字")
	flag.BoolVar(&showVer, "v", false, "打印版本信息")

	// 解析命令行参数
	flag.Parse()

	if showVer {
		version()
		os.Exit(0)
	}

	fmt.Printf("载入配置: %s\n", iniPath)

	// 加载配置文件
	config, err := LoadConfig(iniPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  载入失败: %v\n", err)

		config = &Config{
			InputPath:  defaultInputPath,
			OutputPath: defaultOutputPath,
			NamePrefix: defaultNamePrefix,
			NameFormat: defaultNameFormat,
			StartNum:   defaultStartNum,
		}

		fmt.Fprintf(os.Stderr, "使用默认配置: \n")
		fmt.Fprintf(os.Stderr, "  input: %s\n", config.InputPath)
		fmt.Fprintf(os.Stderr, "  output: %s\n", config.OutputPath)
		fmt.Fprintf(os.Stderr, "  prefix: %s\n", config.NamePrefix)
		fmt.Fprintf(os.Stderr, "  format: %s\n", config.NameFormat)
		fmt.Fprintf(os.Stderr, "  start: %d\n", config.StartNum)
	}

	// 命令行参数优先于配置文件
	if inputPath != "" && inputPath != defaultInputPath {
		fmt.Fprintf(os.Stderr, "手动参数覆盖: input\n")
		fmt.Fprintf(os.Stderr, "  配置值: %s\n", config.InputPath)
		fmt.Fprintf(os.Stderr, "  手动值: %s\n", inputPath)
		config.InputPath = inputPath
	}
	if outputPath != "" && outputPath != defaultOutputPath {
		fmt.Fprintf(os.Stderr, "手动参数覆盖: output\n")
		fmt.Fprintf(os.Stderr, "  配置值: %s\n", config.OutputPath)
		fmt.Fprintf(os.Stderr, "  手动值: %s\n", outputPath)
		config.OutputPath = outputPath
	}
	if namePrefix != "" && namePrefix != defaultNamePrefix {
		fmt.Fprintf(os.Stderr, "手动参数覆盖: prefix\n")
		fmt.Fprintf(os.Stderr, "  配置值: %s\n", config.NamePrefix)
		fmt.Fprintf(os.Stderr, "  手动值: %s\n", namePrefix)
		config.NamePrefix = namePrefix
	}
	if nameFormat != "" && nameFormat != defaultNameFormat {
		fmt.Fprintf(os.Stderr, "手动参数覆盖: format\n")
		fmt.Fprintf(os.Stderr, "  配置值: %s\n", config.NameFormat)
		fmt.Fprintf(os.Stderr, "  手动值: %s\n", nameFormat)
		config.NameFormat = nameFormat
	}
	if startNum > 0 && startNum != defaultStartNum {
		fmt.Fprintf(os.Stderr, "手动参数覆盖: start\n")
		fmt.Fprintf(os.Stderr, "  配置值: %d\n", config.StartNum)
		fmt.Fprintf(os.Stderr, "  手动值: %d\n", startNum)
		config.StartNum = startNum
	}

	// 检查输入路径是文件还是目录
	inputInfo, err := os.Stat(config.InputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法访问输入路径 %s: %v\n", config.InputPath, err)
		return
	}

	fmt.Printf("开始处理: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("  扫描目录(或文件): %s\n", config.InputPath)
	fmt.Printf("  保存目录: %s\n", config.OutputPath)
	fmt.Printf("  起始数字: %d\n", config.StartNum)
	fmt.Printf("  名字前缀: %s\n", config.NamePrefix)
	fmt.Printf("  替换规则: %s\n", config.NameFormat)

	if inputInfo.IsDir() {
		// 如果是目录, 递归处理目录下所有CSV文件
		err = processDirectoryRecursive(config.InputPath, config.OutputPath, config.NamePrefix, config.NameFormat, config.StartNum)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 处理目录失败: %v\n", err)
			return
		}
	} else {
		// 如果是文件, 直接处理该文件
		if !strings.HasSuffix(strings.ToLower(inputInfo.Name()), ".csv") {
			fmt.Fprintf(os.Stderr, "错误: 输入文件不是CSV格式: %s\n", config.InputPath)
			return
		}

		err = processFile(config.InputPath, config.OutputPath, config.NamePrefix, config.NameFormat, config.StartNum)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 处理文件失败: %v\n", err)
			return
		}
	}

	waitForExit()
}

func version() {
	info := fmt.Sprintf(`Server: %s
Version: %s
Go version: %s
Git commit: %s
Built: %s
OS/Arch: %s/%s
User: %s`,
		BuildName, BuildVersion, BuildGoVersion,
		BuildGitCommit, BuildTime, BuildOsName,
		BuildArchName, BuildUser)
	fmt.Println(info)
}

func waitForExit() {
	if runtime.GOOS == "windows" {
		fmt.Println("\n按回车键退出...")
		fmt.Scanln()
	}
}

func LoadConfig(name string) (*Config, error) {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", name)
	}

	data, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("无法读取配置文件: %v", err)
	}

	// 处理UTF-8 BOM
	content := strings.TrimPrefix(string(data), "\xEF\xBB\xBF")

	lines := strings.Split(content, "\n")

	var founded bool
	var currentSection string
	config := &Config{}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 匹配 section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			continue
		}

		// 只处理 rename 区块
		if currentSection == "rename" {
			founded = true

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "input":
				config.InputPath = value
			case "output":
				config.OutputPath = value
			case "prefix":
				config.NamePrefix = value
			case "format":
				config.NameFormat = value
			case "start":
				fmt.Sscanf(value, "%d", &config.StartNum)
			}
		}
	}

	if !founded {
		return nil, fmt.Errorf("未找到rename配置信息")
	}

	fmt.Printf("找到rename配置: \n")
	if config.InputPath == "" {
		config.InputPath = defaultInputPath
		fmt.Printf("  input未设置, 使用默认值: %s\n", config.InputPath)
	} else {
		fmt.Printf("  input: √\n")
	}

	if config.OutputPath == "" {
		config.OutputPath = defaultOutputPath
		fmt.Printf("  output未设置, 使用默认值: %s\n", config.OutputPath)
	} else {
		fmt.Printf("  output: √\n")
	}

	if config.NamePrefix == "" {
		config.NamePrefix = defaultNamePrefix
		fmt.Printf("  prefix未设置, 使用默认值: %s\n", config.NamePrefix)
	} else {
		fmt.Printf("  prefix: √\n")
	}

	if config.NameFormat == "" {
		config.NameFormat = defaultNameFormat
		fmt.Printf("  format未设置, 使用默认值: %s\n", config.NameFormat)
	} else {
		fmt.Printf("  format: √\n")
	}

	if config.StartNum <= 0 {
		config.StartNum = defaultStartNum
		fmt.Printf("  start未设置(或无效), 使用默认值: %d\n", config.StartNum)
	} else {
		fmt.Printf("  start: √\n")
	}

	return config, nil
}

// processDirectoryRecursive 递归处理目录下所有CSV文件，保持目录结构
func processDirectoryRecursive(inputDir, outputDir, namePrefix, nameFormat string, startNum int) error {
	counter := startNum
	processedCount := 0

	// 使用filepath.WalkDir递归遍历目录
	err := filepath.WalkDir(inputDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if d.IsDir() {
			return nil
		}

		// 只处理CSV文件
		if strings.HasSuffix(strings.ToLower(d.Name()), ".csv") {
			// 计算相对于输入目录的路径
			relPath, err := filepath.Rel(inputDir, path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "警告: 无法计算相对路径 %s: %v\n", path, err)
				return nil
			}

			// 计算输出文件路径
			outputFilePath := filepath.Join(outputDir, relPath)

			// 处理单个文件
			count, err := processSingleFile(path, outputFilePath, namePrefix, nameFormat, counter)
			if err != nil {
				fmt.Fprintf(os.Stderr, "警告: 处理文件 %s 失败: %v\n", path, err)
				return nil
			}

			// 更新计数器
			counter += count
			processedCount++

			// 显示相对路径
			fmt.Printf("已处理文件: %s, 更新了 %d 条记录\n", relPath, count)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("遍历目录时出错: %v", err)
	}

	fmt.Printf("结束处理: \n")

	if processedCount > 0 {
		fmt.Printf("  处理CSV文件数量: %d\n", processedCount)
		fmt.Printf("  更新记录数量: %d\n", counter-startNum)
		fmt.Printf("  最终编号数字: %d\n", counter-1)
	} else {
		fmt.Println("  未找到任何CSV文件进行处理")
	}

	return nil
}

// processFile 处理单个CSV文件
func processFile(inputFile, outputFile, namePrefix, nameFormat string, startNum int) error {
	// 如果输出路径是目录, 则在该目录下创建同名文件
	if info, err := os.Stat(outputFile); err == nil && info.IsDir() {
		_, fileName := filepath.Split(inputFile)
		outputFile = filepath.Join(outputFile, fileName)
	} else {
		// 确保输出文件的目录存在
		outputDir := filepath.Dir(outputFile)
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			err = os.MkdirAll(outputDir, 0755)
			if err != nil {
				return fmt.Errorf("无法创建输出目录 %s: %v", outputDir, err)
			}
		}
	}

	count, err := processSingleFile(inputFile, outputFile, namePrefix, nameFormat, startNum)
	if err != nil {
		return err
	}

	fmt.Printf("成功处理文件: %s, 更新了 %d 条记录\n", inputFile, count)
	return nil
}

// processSingleFile 处理单个CSV文件的核心逻辑
func processSingleFile(inputFile, outputFile, namePrefix, nameFormat string, startNum int) (int, error) {
	// 读取CSV文件
	file, err := os.Open(inputFile)
	if err != nil {
		return 0, fmt.Errorf("无法打开文件 %s: %v", inputFile, err)
	}
	defer file.Close()

	// 读取CSV记录
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("无法读取CSV内容: %v", err)
	}

	// 检查是否有数据
	if len(records) <= 1 {
		return 0, fmt.Errorf("CSV文件中没有足够的数据")
	}

	// 查找name字段的索引
	headers := records[0]
	nameIndex := -1
	for i, header := range headers {
		if header == "name" {
			nameIndex = i
			break
		}
	}

	if nameIndex == -1 {
		return 0, fmt.Errorf("CSV文件中未找到name字段")
	}

	// 更新name字段
	counter := startNum
	for i := 1; i < len(records); i++ {
		records[i][nameIndex] = formatName(nameFormat, namePrefix, counter)
		counter++
	}

	// 确保输出文件的目录存在
	outputDir := filepath.Dir(outputFile)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			return 0, fmt.Errorf("无法创建输出目录 %s: %v", outputDir, err)
		}
	}

	// 写回文件
	output, err := os.Create(outputFile)
	if err != nil {
		return 0, fmt.Errorf("无法创建文件 %s: %v", outputFile, err)
	}
	defer output.Close()

	writer := csv.NewWriter(output)
	defer writer.Flush()

	err = writer.WriteAll(records)
	if err != nil {
		return 0, fmt.Errorf("无法写入CSV内容: %v", err)
	}

	// 返回处理的记录数
	return len(records) - 1, nil
}

// formatName 根据给定的格式字符串和参数生成新名称
// {0} 表示原始名称，{1} 表示序号
func formatName(format, prefix string, number int) string {
	result := format
	result = strings.ReplaceAll(result, "{0}", prefix)
	result = strings.ReplaceAll(result, "{1}", fmt.Sprintf("%d", number))
	return result
}
