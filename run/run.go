package run

import (
	"CalcNMR/calc"
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
	filename      string
	version       bool
	help          bool
	config        string
	skip          string
	preThreshold  float64
	postThreshold float64
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
	// 新建 config 参数，代表配置 toml 文件，如果为空，则读取当前目录下的 config.ini
	flag.StringVar(&c.config, "config", "", "specify configuration file path")
	// 新建一个 skip 参数，代表是否跳过某一步骤
	flag.StringVar(&c.skip, "skip", "", "specify a step to skip (e.g., 0: md; 1: pre-opt; 2: post-opt)")
	// 新建一个 prethre 阈值参数，用来判断做完预优化后调用 doublecheck 的阈值
	flag.Float64Var(&c.preThreshold, "prethre", 0.2, "threshold for invoking doubleCheck after pre-optimization")
	// 新建一个 postthre 阈值参数，用来判断做完进一步优化后调用 doublecheck 的阈值
	flag.Float64Var(&c.postThreshold, "postthre", 0.2, "threshold for invoking doubleCheck after post-optimization")
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
Usage: CalcNMR <input> [options]

Input: Files with atomic coordinates (e.g., xyz files)

Options:
  --version     Display version
  --help        Display help information
  --config      Specify configuration file path
  --skip        Specify a step to skip (e.g., 0: md; 1: pre-opt; 2: post-opt)	
  --prethre     threshold for invoking doubleCheck after pre-optimization
  --postthre    threshold for invoking doubleCheck after post-optimization
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

	// 如果为空，则读取当前运行脚本的目录下的 config.ini
	// 如果当前目录下不存在 config.ini 则报错
	// 如果不为空，则读取目标文件，同时需要判断输入的 ini 文件是否存在，如果存在，打印读取成功
	// 如果不存在，则打印错误
	if c.config == "" {
		checkConfig, configFullPath := utils.CheckFileCurrentExist("config.ini")
		if checkConfig {
			c.config = configFullPath
			fmt.Println("Hint: Successfully read the configuration file path: " + configFullPath)
		} else {
			fmt.Println("Error: The default configuration file was not found in the current directory: config.ini")
			fmt.Println("Hint: Please specify the configuration file path.")
			os.Exit(1)
		}
	} else {
		// 将 c.config 转化为一个绝对路径
		configFullPath, err := filepath.Abs(c.config)
		if err != nil {
			fmt.Println("Error getting absolute path:", err)
		}
		// 检查输入的 config 是否为一个 ini 文件
		checkToml := utils.CheckFileType(configFullPath, ".ini")
		if checkToml {
			fmt.Println("Hint: Successfully read the configuration file path: " + configFullPath)
		} else {
			fmt.Println("Error: Please enter a toml type configuration file.")
			os.Exit(1)
		}
	}

	fmt.Println()
	// ----------------------------------------------------------------
	// 开始运行 xtb 程序做动力学模拟
	// ----------------------------------------------------------------
	dyConfig := calc.ParseConfigFile(c.config).DyConfig
	if c.skip != "0" {
		fmt.Println("Running xtb for dynamics simulation...")
		calc.XtbExecuteMD(&dyConfig, c.filename)
	}

	fmt.Println()
	// ----------------------------------------------------------------
	// 开始运行 crest 程序做预优化
	// ----------------------------------------------------------------
	optConfig := calc.ParseConfigFile(c.config).OptConfig
	if c.skip != "1" {
		fmt.Println("Running crest for pre-optimization...")
		calc.XtbExecutePreOpt(&optConfig, "dynamics.xyz")
		// ----------------------------------------------------------------
		// 对 crest 预优化产生的 pre-optimization 文件进行 DoubleCheck
		// ----------------------------------------------------------------
		// 读取生成的 pre_opt.xyz 文件
		preClusters, err := calc.ParseXyzFile("pre_opt.xyz")
		if err != nil {
			fmt.Println("Error Parse xyz file:", err)
			return
		}
		calc.DoubleCheck(c.preThreshold, preClusters)
	}

	// ----------------------------------------------------------------
	// 开始运行 crest 程序做进一步优化
	// ----------------------------------------------------------------
	if c.skip != "2" {
		fmt.Println("Running crest for post-optimization...")
		calc.XtbExecutePostOpt(&optConfig, "pre_cluster.xyz")
		// ----------------------------------------------------------------
		// 对 crest 进一步产生的 post-optimization 文件进行 DoubleCheck
		// ----------------------------------------------------------------
	}

}
