package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

type Item struct {
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	Image         string      `json:"image"`
	Edition       int         `json:"edition"`
	Attributes    []Attribute `json:"attributes"`
	ParentEdition string      `json:"parent_edition"`
}

type Attribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}

type AttributesMap map[string]map[string]string

func main() {
	var iniPath, outPath string
	var showVer bool
	flag.StringVar(&iniPath, "c", "convert_okx.ini", "ini配置文件路径")
	flag.StringVar(&outPath, "o", "okx.csv", "导出的 OKX CSV 文件路径")
	flag.BoolVar(&showVer, "v", false, "打印版本信息")
	flag.Parse()

	if showVer {
		version()
		os.Exit(0)
	}

	if len(iniPath) == 0 {
		fmt.Println("缺少 -c 参数, 使用 -h 可查看帮助")
		return
	}

	config, err := LoadConfig(iniPath)
	if err != nil {
		panic(err)
	}

	// 读取 2/_metadata.json
	fileData2, err := os.ReadFile(config.File2Path)
	if err != nil {
		panic(err)
	}

	var items2 []Item
	err = json.Unmarshal(fileData2, &items2)
	if err != nil {
		panic(err)
	}

	// 读取 1/_metadata.json
	fileData1, err := os.ReadFile(config.File1Path)
	if err != nil {
		panic(err)
	}

	var items1 []Item
	err = json.Unmarshal(fileData1, &items1)
	if err != nil {
		panic(err)
	}

	// 针对第一批结果, 构建 edition 到 attributes 的映射表
	editionToAttributesMap := make(AttributesMap)
	for _, item := range items1 {
		editionStr := fmt.Sprintf("%d", item.Edition)
		editionToAttributesMap[editionStr] = make(map[string]string)
		for _, v := range item.Attributes {
			editionToAttributesMap[editionStr][v.TraitType] = v.Value
		}
	}

	// 构建数据行并收集所有列名（静态 + 动态）
	var rows []map[string]string
	columnsMap := make(map[string]struct{})

	var key = ""
	for _, item := range items2 {
		row := map[string]string{
			"name":        item.Name,
			"description": item.Description,
			"file_name":   getLastPart(item.Image),
		}

		// 合并当前批次的特征数据
		for _, attr := range item.Attributes {
			// 提取上级元数据标识
			if attr.TraitType == config.ParentKey {
				item.ParentEdition = attr.Value
				break
			} else {
				key = getAttributeRowTitle(attr.TraitType)
				row[key] = attr.Value
				columnsMap[key] = struct{}{}
				fmt.Printf("key: %s, value: %s\n", key, row[key])
			}
		}

		// 按标识查找合并上级元数据
		if parentAttrs, ok := editionToAttributesMap[item.ParentEdition]; ok {
			for k, v := range parentAttrs {
				key = getAttributeRowTitle(k)

				// 上级元数据优先级低, 可以被下级元数据覆盖
				if _, ok := row[key]; ok {
					continue
				}

				row[key] = v
				columnsMap[key] = struct{}{}
			}
		}

		rows = append(rows, row)
	}

	// 提取所有列名（静态字段 + 动态属性）
	columns := []string{"name", "description", "file_name"}
	for col := range columnsMap {
		columns = append(columns, col)
	}

	// 写入CSV文件
	file, err := os.Create(outPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	err = writer.Write(columns)
	if err != nil {
		panic(err)
	}

	for _, row := range rows {
		var record []string
		for _, col := range columns {
			record = append(record, row[col])
		}
		err = writer.Write(record)
		if err != nil {
			panic(err)
		}
	}
	writer.Flush()
}

func version() {
	_, _ = fmt.Fprintf(os.Stdout, "Server: %s\n", BuildName)
	_, _ = fmt.Fprintf(os.Stdout, " Version: %s\n", BuildVersion)
	_, _ = fmt.Fprintf(os.Stdout, " Go version: %s\n", BuildGoVersion)
	_, _ = fmt.Fprintf(os.Stdout, " Git commit: %s\n", BuildGitCommit)
	_, _ = fmt.Fprintf(os.Stdout, " Built: %s\n", BuildTime)
	_, _ = fmt.Fprintf(os.Stdout, " OS/Arch: %s/%s\n", BuildOsName, BuildArchName)
	_, _ = fmt.Fprintf(os.Stdout, " User: %s\n", BuildUser)
}

// 获取路径的最后一部分作为文件名
func getLastPart(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

// 获取特征值列标题
func getAttributeRowTitle(traitType string) string {
	return fmt.Sprintf("attributes[%s]", traitType)
}

// Config 保存配置信息
type Config struct {
	File1Path string
	File2Path string
	ParentKey string
}

// LoadConfig 从 convert_okx.ini 中加载配置
func LoadConfig(name string) (*Config, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("无法读取配置文件: %v", err)
	}

	lines := strings.Split(string(data), "\n")

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

		// 只处理 paths 和 fields 区块
		if currentSection == "paths" || currentSection == "fields" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch {
			case currentSection == "paths" && key == "file1":
				config.File1Path = value
			case currentSection == "paths" && key == "file2":
				config.File2Path = value
			case currentSection == "fields" && key == "parent_key":
				config.ParentKey = value
			}
		}
	}

	if config.File1Path == "" || config.File2Path == "" || config.ParentKey == "" {
		return nil, fmt.Errorf("配置缺失必要字段")
	}

	return config, nil
}
