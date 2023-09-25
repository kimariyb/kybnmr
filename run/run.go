package run

import (
	"flag"
	"fmt"
	"kybnmr/calc"
	"kybnmr/utils"
	"os"
	"path/filepath"
)

/*
* run.go
* 该模块用来处理如何在命令行中使用参数运行 KYBNMR 程序
*
* @Struct:
*	KYBNMR: 运行 KYBNMR 所需要的参数集合，封装为一个结构体
*
* @Method:
*	NewKYBNMR(): 新建一个 KYBNMR 对象
*		Return: *KYBNMR
*
*	ParseArgs(): 解析 KYBNMR 运行所需要的参数
*
*	ShowHelp(): 展示 Help 参数，打印运行 KYBNMR 所需要的参数信息
*
*	Run(): 起到通过命令行执行整个任务流程的作用
*
* @Author: Kimariyb
* @Address: XiaMen University
* @Data: 2023-09-21
 */

type KYBNMR struct {
	filename string
	version  bool
	help     bool
	config   string
	md       int
	pre      int
	post     int
	opt      int
	sp       int
}

func NewKYBNMR() *KYBNMR {
	return &KYBNMR{}
}

// ParseArgs 解析 KYBNMR 在命令行中运行所使用的参数
func (c *KYBNMR) ParseArgs() {
	// 新建 version 参数，代表版权信息，默认不显示
	flag.BoolVar(&c.version, "version", false, "display version")
	// 新建 help 参数，代表帮助选项，默认不显示
	flag.BoolVar(&c.help, "help", false, "display help information")
	// 新建 config 参数，代表配置 toml 文件，如果为空，则读取当前目录下的 config.ini
	flag.StringVar(&c.config, "config", "", "specify configuration file path")
	// 新建一个 opt 参数，代表做优化和振动分析时使用什么程序，默认为 0: gaussian
	flag.IntVar(&c.opt, "opt", 0, "Select the program to be used when doing DFT optimization and vibration analysis (e.g., 0: Gaussian; 1: Orca)")
	// 新建一个 sp 参数，代表做单点时使用什么程序，默认为 1: orca
	flag.IntVar(&c.sp, "sp", 1, "Select the program to be used when doing DFT single point (e.g., 0: Gaussian; 1: Orca)")
	// 新建一个 md 参数，代表是否开启动力学模拟步骤
	flag.IntVar(&c.md, "md", 1, "Whether molecular dynamics simulations are performed (0: false; 1: true [Default])")
	// 新建一个 pre 参数，代表是是否开启 xtb/crest 预优化步骤
	flag.IntVar(&c.pre, "pre", 1, "whether to use crest for post-optimization (0: false; 1: true [Default])")
	// 新建一个 post 参数，代表是否开启 xtb/crest 做进一步优化
	flag.IntVar(&c.post, "post", 1, "whether to use crest for post-optimization (0: false; 1: true [Default])")

	// 解析参数
	flag.Parse()

	// 将 post，pre，md, opt, sp参数的值限制在 0 和 1 之间
	if c.post != 0 && c.post != 1 {
		fmt.Println("Warning: The parameter --post must be one of 0 and 1 !")
		fmt.Println("If you entered a parameter other than 0 or 1, the parameter has now been adjusted to 1.")
		c.post = 1
	}
	if c.pre != 0 && c.pre != 1 {
		fmt.Println("Warning: The parameter --pre must be one of 0 and 1 !")
		fmt.Println("If you entered a parameter other than 0 or 1, the parameter has now been adjusted to 1.")
		c.pre = 1
	}
	if c.md != 0 && c.md != 1 {
		fmt.Println("Warning: The parameter --md must be one of 0 and 1 !")
		fmt.Println("If you entered a parameter other than 0 or 1, the parameter has now been adjusted to 1.")
		c.md = 1
	}
	if c.opt != 0 && c.opt != 1 {
		fmt.Println("Warning: The parameter --opt must be one of 0 and 1 !")
		fmt.Println("If you entered a parameter other than 0 or 1, the parameter has now been adjusted to 0.")
		c.opt = 0
	}
	if c.sp != 0 && c.sp != 1 {
		fmt.Println("Warning: The parameter --sp must be one of 0 and 1 !")
		fmt.Println("If you entered a parameter other than 0 or 1, the parameter has now been adjusted to 1.")
		c.sp = 1
	}
}

// ShowHelp 展示 Help 参数
func (c *KYBNMR) ShowHelp() {
	helpText := `
Usage: KYBNMR <input> [options]

Input: Files with atomic coordinates (e.g., xyz files)

Options:
  --version     Display version
  --help        Display help information
  --config      Specify configuration file path
  --opt         Select the program to be used when doing DFT optimization
                (e.g., 0: Gaussian [Default]; 1: Orca)
  --sp          Select the program to be used when doing DFT single point 
                (e.g., 0: Gaussian; 1: Orca [Default])
  --md          Whether molecular dynamics simulations are performed 
                (0: false; 1: true [Default])
  --pre         Whether to use crest for pre-optimization
                (0: false; 1: true [Default])
  --post        Whether to use crest for post-optimization
                (0: false; 1: true [Default])
`
	fmt.Println(helpText)
}

// Run 起到通过命令行执行整个任务流程的作用
func (c *KYBNMR) Run() {
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

	// 获取配置信息
	optConfig := calc.ParseConfigFile(c.config).OptConfig
	dyConfig := calc.ParseConfigFile(c.config).DyConfig

	// ----------------------------------------------------------------
	// 开始运行 xtb 程序做动力学模拟
	// ----------------------------------------------------------------
	if c.md == 1 {
		// 如果 c.md 为 false，执行动力学步骤
		// runMolecularDynamics()
		fmt.Println()
		fmt.Println("Running xtb for dynamics simulation...")
		calc.XtbExecuteMD(&dyConfig, c.filename)
	}
	// ----------------------------------------------------------------
	// 开始运行 crest 程序做预优化
	// ----------------------------------------------------------------
	if c.pre == 1 {
		// 如果 c.pre 为 false，执行预优化步骤
		// runPreOptimization()
		fmt.Println()
		fmt.Println("Running crest for pre-optimization...")
		calc.XtbExecutePreOpt(&optConfig, "dynamics.xyz")
		// 对 crest 预优化产生的 pre-optimization 文件进行 DoubleCheck
		// 读取生成的 pre_opt.xyz 文件
		preClusters, err := calc.ParseXyzFile("pre_opt.xyz")
		if err != nil {
			fmt.Println("Error Parse xyz file:", err)
			return
		}
		// 获取 doublecheck 阈值
		preThreshold := utils.SplitStringByComma(optConfig.PreThreshold)
		// 进行 double check，同时得到 clusters
		preRemainClusters, err := calc.DoubleCheck(preThreshold[0], preThreshold[1], preClusters)
		if err != nil {
			fmt.Println("Error Running DoubleCheck", err)
			return
		}
		// 写入到新的 xyz 文件中
		calc.WriteToXyzFile(preRemainClusters, "pre_clusters.xyz")
	}
	// ----------------------------------------------------------------
	// 开始运行 crest 程序做进一步优化
	// ----------------------------------------------------------------
	if c.post == 1 {
		// 如果 c.post 为 false，执行进一步优化步骤
		// runFurtherOptimization()
		fmt.Println("Running crest for post-optimization...")
		calc.XtbExecutePostOpt(&optConfig, "pre_clusters.xyz")
		// 对 crest 进一步产生的 post-optimization 文件进行 DoubleCheck
		// 读取生成的 post_opt.xyz 文件
		postClusters, err := calc.ParseXyzFile("post_opt.xyz")
		if err != nil {
			fmt.Println("Error Parse xyz file:", err)
			return
		}
		// 获取 doublecheck 阈值
		postThreshold := utils.SplitStringByComma(optConfig.PostThreshold)
		// 进行 double check，同时得到 clusters
		postRemainClusters, err := calc.DoubleCheck(postThreshold[0], postThreshold[1], postClusters)
		if err != nil {
			fmt.Println("Error Running DoubleCheck", err)
			return
		}
		// 写入到新的 xyz 文件中
		calc.WriteToXyzFile(postRemainClusters, "post_clusters.xyz")
	}
	// ----------------------------------------------------------------
	// 开始运行 gaussian/orca 程序做 DFT 优化
	// ----------------------------------------------------------------
	// 执行 DFT 步骤
	postRemainClusters, err := calc.ParseXyzFile("post_clusters.xyz")
	if err != nil {
		fmt.Println("Error Parse xyz file:", err)
		return
	}
	// runDFT()
	fmt.Println("Running Gaussian for DFT Optimization Calculating...")
	if c.opt == 0 {
		// 运行 Gaussian 程序优化结构
		calc.ExecuteOptimization(optConfig.GauPath, "GauTemplate.gjf", postRemainClusters, "gaussian")
	} else if c.opt == 1 {
		// 运行 Orca 程序优化结构
		calc.ExecuteOptimization(optConfig.OrcaPath, "OrcaTemplate.inp", postRemainClusters, "orca")
	}

}
