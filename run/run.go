package run

import (
	"CalcNMR/utils"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

/*
* run.go
* 该模块用来处理如何在命令行中使用参数运行 CalcNMR 程序
*
* @Struct:
*	CalcNMR: 运行 CalcNMR 所需要的参数集合，封装为一个结构体
*
* @Method:
*	NewCalcNMR(): 新建一个 CalcNMR 对象
*		Return: *CalcNMR
*
*	ParseArgs(): 解析 CalcNMR 运行所需要的参数
*
*	ShowHelp(): 展示 Help 参数，打印运行 CalcNMR 所需要的参数信息
*
*	Run(): 起到通过命令行执行整个任务流程的作用
*
* @Author: Kimariyb
* @Address: XiaMen University
* @Data: 2023-09-21
 */

type CalcNMR struct {
	filename string
	version  bool
	help     bool
	config   string
}

func NewCalcNMR() *CalcNMR {
	return &CalcNMR{}
}

// ParseArgs 解析 CalcNMR 在命令行中运行所使用的参数
func (c *CalcNMR) ParseArgs() {
	// 新建 version 参数，代表版权信息，默认不显示
	flag.BoolVar(&c.version, "version", false, "display version")
	// 新建 help 参数，代表帮助选项，默认不显示
	flag.BoolVar(&c.help, "help", false, "display help information")
	// 新建 config 参数，代表配置 toml 文件，如果为空，则读取当前目录下的 config.toml
	flag.StringVar(&c.config, "config", "", "specify configuration file path")
	// 解析参数
	flag.Parse()

	// 获取运行的参数
	args := flag.Args()
	// 如果参数字段大于 0 将第一个元素赋值为 filename
	if len(args) > 0 {
		c.filename = args[0]
	}
}

// ShowHelp 展示 Help 参数
func (c *CalcNMR) ShowHelp() {
	helpText := `
Usage: cnmr <input> [options]

Input: Files with atomic coordinates (e.g., xyz files)

Options:
  --version     Display version
  --help        Display help information
  --config      Specify configuration file path
`
	fmt.Println(helpText)
}

// Run 起到通过命令行执行整个任务流程的作用
func (c *CalcNMR) Run() {
	if c.version {
		utils.ShowHead()
		os.Exit(0)
	}

	if c.help {
		c.ShowHelp()
		os.Exit(0)
	}

	// 展示程序的基础信息、版本信息以及作者信息
	utils.ShowHead()

	// 首先判断输入的 filename 是否为空，如果为空，则直接打印错误
	// 接着判断传入的 filename 是否为一个 xyz 文件，xyz 文件是一个记录分子原子信息的文件
	// 如果传入的是一个 xyz 文件，但是没有扫描到，则报错。
	if c.filename == "" {
		fmt.Println("Error: Please provide an input filename!")
		os.Exit(1)
	} else {
		// 将 filename 转变为一个绝对路径
		inputFullPath, err := filepath.Abs(c.filename)
		if err != nil {
			fmt.Println("Error getting absolute path:", err)
		}
		// 检查输入的 filename 是否为一个 xyz 文件
		checkXyz := utils.CheckFileType(inputFullPath, ".xyz")
		if checkXyz {
			fmt.Println("Hint: Successfully read the input file path: " + inputFullPath)
		} else {
			fmt.Println("Error: Please enter an input file of type xyz.")
			os.Exit(1)
		}

	}

	// 如果为空，则读取当前运行脚本的目录下的 config.toml
	// 如果当前目录下不存在 config.toml 则报错
	// 如果不为空，则读取目标文件，同时需要判断输入的 toml 文件是否存在，如果存在，打印读取成功
	// 如果不存在，则打印错误
	if c.config == "" {
		checkConfig, configFullPath := utils.CheckFileCurrentExist("config.toml")
		if checkConfig {
			c.config = configFullPath
			fmt.Println("Hint: Successfully read the configuration file path: " + configFullPath)
		} else {
			fmt.Println("Error: The default configuration file was not found in the current directory: config.toml")
			fmt.Println("Hint: Please specify the configuration file path.")
			os.Exit(1)
		}
	} else {
		// 将 c.config 转化为一个绝对路径
		configFullPath, err := filepath.Abs(c.config)
		if err != nil {
			fmt.Println("Error getting absolute path:", err)
		}
		// 检查输入的 config 是否为一个 toml 文件
		checkToml := utils.CheckFileType(configFullPath, ".toml")
		if checkToml {
			fmt.Println("Hint: Successfully read the configuration file path: " + configFullPath)
		} else {
			fmt.Println("Error: Please enter a toml type configuration file.")
			os.Exit(1)
		}
	}

	// 开始运行 xtb 程序

}