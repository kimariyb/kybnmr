package run

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"kybnmr/calc"
	"kybnmr/utils"
	"log"
	"os"
	"path/filepath"
	"time"
)

/*
* run.go
* 该模块用来处理如何在命令行中使用参数运行 KYBNMR 程序
*
* @Author: Kimariyb
* @Address: XiaMen University
* @Data: 2023-09-21
 */

type KYBNMR struct {
	input   string
	version bool
	help    bool
	config  string
	md      IsOpenOption
	pre     IsOpenOption
	post    IsOpenOption
	opt     DFTOption
	sp      DFTOption
}

type IsOpenOption int

const (
	OpenFalse IsOpenOption = 0
	OpenTure  IsOpenOption = 1
)

type DFTOption int

const (
	DFTGaussian DFTOption = 0
	DFTOrca     DFTOption = 1
)

func NewKYBNMR() *KYBNMR {
	return &KYBNMR{}
}

// 首先判断输入的文件是否为空，如果为空，则直接打印错误
// 接着判断传入的文件是否为一个 xyz 文件，xyz 文件是一个记录分子原子信息的文件
// 如果传入的是一个 xyz 文件，但是没有扫描到，则报错。
func (k *KYBNMR) checkInputFile() error {
	if k.input == "" {
		return fmt.Errorf("error: please provide an input filename")
	}

	inputFullPath, err := filepath.Abs(k.input)
	if err != nil {
		return fmt.Errorf("error getting absolute path: %w", err)
	}

	checkXyz := utils.CheckFileType(inputFullPath, ".xyz")
	if !checkXyz {
		return fmt.Errorf("error: please enter an input file of type xyz")
	}

	fmt.Println("Hint: Successfully read the input file path: " + inputFullPath)
	return nil
}

// 如果为空，则读取当前运行脚本的目录下的 config.ini
// 如果当前目录下不存在 config.ini 则报错
// 如果不为空，则读取目标文件，同时需要判断输入的 ini 文件是否存在，如果存在，打印读取成功
// 如果不存在，则打印错误
func (k *KYBNMR) checkConfigFile() error {
	if k.config == "" {
		checkConfig, configFullPath := utils.CheckFileCurrentExist("config.ini")
		if !checkConfig {
			return fmt.Errorf("error: the default configuration file was not found in the current directory: config.ini")
		}
		k.config = configFullPath
		fmt.Println("Hint: Successfully read the config file path: " + configFullPath)
	}

	return nil
}

func (k *KYBNMR) runPreOptimization(optConfig *calc.OptimizedConfig) error {
	calc.XtbExecutePreOpt(optConfig, "dynamics.xyz")
	// 对 crest 预优化产生的 pre-optimization 文件进行 DoubleCheck
	// 读取生成的 pre_opt.xyz 文件
	preClusters, err := calc.ParseXyzFile("pre_opt.xyz")
	if err != nil {
		fmt.Println("Error Parse xyz file:", err)
		return nil
	}
	// 获取 doublecheck 阈值
	preThreshold := utils.SplitStringByComma(optConfig.PreThreshold)
	// 进行 double check，同时得到 clusters
	preRemainClusters, err := calc.DoubleCheck(preThreshold[0], preThreshold[1], preClusters)
	if err != nil {
		fmt.Println("Error Running DoubleCheck", err)
		return nil
	}
	// 写入到新的 xyz 文件中
	calc.WriteToXyzFile(preRemainClusters, "pre_clusters.xyz")
	return nil
}

func (k *KYBNMR) runFurtherOptimization(optConfig *calc.OptimizedConfig) error {
	fmt.Println("Running crest for post-optimization...")
	calc.XtbExecutePostOpt(optConfig, "pre_clusters.xyz")
	// 对 crest 进一步产生的 post-optimization 文件进行 DoubleCheck
	// 读取生成的 post_opt.xyz 文件
	postClusters, err := calc.ParseXyzFile("post_opt.xyz")
	if err != nil {
		fmt.Println("Error Parse xyz file:", err)
		return nil
	}
	// 获取 doublecheck 阈值
	postThreshold := utils.SplitStringByComma(optConfig.PostThreshold)
	// 进行 double check，同时得到 clusters
	postRemainClusters, err := calc.DoubleCheck(postThreshold[0], postThreshold[1], postClusters)
	if err != nil {
		fmt.Println("Error Running DoubleCheck", err)
		return nil
	}
	// 写入到新的 xyz 文件中
	calc.WriteToXyzFile(postRemainClusters, "post_clusters.xyz")
	return nil
}

func (k *KYBNMR) ParseArgsToRun() {
	// EXAMPLE: Override a template
	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[OPTIONS]{{end}}{{if .Commands}} [command] <input> {{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
`
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "print only the version",
	}
	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "show help",
	}
	app := &cli.App{
		Name:    "kybnmr",
		Usage:   "A scripting program for fully automated calculation of NMR of large molecules",
		Version: "v1.0.0(dev)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       "config.ini",
				Usage:       "Load configuration from `FILE`",
				Destination: &k.config,
			},
			&cli.IntFlag{
				Name:        "opt",
				Usage:       "DFT optimization and vibration procedure",
				Aliases:     []string{"o"},
				Destination: (*int)(&k.opt),
				Value:       int(DFTGaussian),
			},
			&cli.IntFlag{
				Name:        "sp",
				Usage:       "DFT single point procedure",
				Aliases:     []string{"s"},
				Destination: (*int)(&k.sp),
				Value:       int(DFTOrca),
			},
			&cli.IntFlag{
				Name:        "md",
				Usage:       "whether molecular dynamics simulations are performed",
				Aliases:     []string{"m"},
				Destination: (*int)(&k.md),
				Value:       int(OpenTure),
			},
			&cli.IntFlag{
				Name:        "pre",
				Usage:       "whether to use crest for pre-optimization",
				Aliases:     []string{"pr"},
				Destination: (*int)(&k.pre),
				Value:       int(OpenTure),
			},
			&cli.IntFlag{
				Name:        "post",
				Usage:       "whether to use crest for post-optimization",
				Aliases:     []string{"po"},
				Destination: (*int)(&k.post),
				Value:       int(OpenTure),
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return fmt.Errorf("missing required argument: <input>")
			}
			k.input = c.Args().Get(0)
			// Run the workflow
			if err := k.Run(); err != nil {
				log.Fatal(err)
			}
			return nil
		},
		Authors: []*cli.Author{
			{
				Name:  "Kimari Y.B.",
				Email: "kimariyb@163.com",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// Run 起到通过命令行执行整个任务流程的作用
func (k *KYBNMR) Run() error {
	// 记录起始时间
	start := time.Now()

	// 展示程序的基础信息、版本信息以及作者信息
	utils.ShowHead()

	if err := k.checkInputFile(); err != nil {
		return err
	}

	if err := k.checkConfigFile(); err != nil {
		return err
	}

	// 获取配置信息
	optConfig := calc.ParseConfigFile(k.config).OptConfig
	dyConfig := calc.ParseConfigFile(k.config).DyConfig
	spConfig := calc.ParseConfigFile(k.config).SpConfig
	// ----------------------------------------------------------------
	// 开始运行 xtb 程序做动力学模拟
	// ----------------------------------------------------------------
	fmt.Println()
	if k.md == OpenTure {
		fmt.Println("Running xtb for dynamics simulation...")
		if err := calc.XtbExecuteMD(&dyConfig, k.input); err != nil {
			return err
		}
	} else if k.md == OpenFalse {
		fmt.Println("Skipped dynamics simulation")
	}
	// ----------------------------------------------------------------
	// 开始运行 crest 程序做预优化
	// ----------------------------------------------------------------
	fmt.Println()
	if k.pre == OpenTure {
		fmt.Println("Running crest for pre-optimization...")
		if err := k.runPreOptimization(&optConfig); err != nil {
			return err
		}
	} else if k.pre == OpenFalse {
		fmt.Println("Skipped pre-optimization")
	}
	// ----------------------------------------------------------------
	// 开始运行 crest 程序做进一步优化
	// ----------------------------------------------------------------
	fmt.Println()
	if k.post == OpenTure {
		fmt.Println("Running crest for post-optimization...")
		if err := k.runFurtherOptimization(&optConfig); err != nil {
			return err
		}
	} else if k.post == OpenFalse {
		fmt.Println("Skipped post-optimization")
	}

	postRemainClusters, err := calc.ParseXyzFile("post_clusters.xyz")
	if err != nil {
		return fmt.Errorf("error parsing xyz file: %w", err)
	}

	// ----------------------------------------------------------------
	// 开始运行 gaussian/orca 程序做 DFT 优化
	// ----------------------------------------------------------------
	// 执行 DFT 步骤
	fmt.Println()
	fmt.Println("Running Gaussian/Orca for DFT Optimization Calculating...")
	if k.opt == DFTGaussian {
		err = calc.RunDFTOptimization(optConfig.GauPath, "GauTemplate.gjf", postRemainClusters, "gaussian")
	} else if k.opt == DFTOrca {
		err = calc.RunDFTOptimization(optConfig.OrcaPath, "OrcaTemplate.inp", postRemainClusters, "orca")
	}
	if err != nil {
		return fmt.Errorf("error running DFT optimization: %w", err)
	}

	// ----------------------------------------------------------------
	// 开始运行 gaussian/orca 程序做 DFT 单点能计算
	// ----------------------------------------------------------------
	// 执行 DFT 步骤
	fmt.Println()
	fmt.Println("Running Gaussian/Orca for DFT Single Point Energy Calculating...")
	if k.sp == DFTGaussian {
		// 调用 Multiwfn 将 out 文件全都转化为 inp 文件或 gjf 文件
		err = calc.BatchMTFToGenerateFile("gaussian", "/thermo/opt", &spConfig)
		err = calc.RunDFTSinglePoint(optConfig.GauPath, "gaussian")
	} else if k.sp == DFTOrca {
		// 调用 Multiwfn 将 out 文件全都转化为 inp 文件或 gjf 文件
		err = calc.BatchMTFToGenerateFile("orca", "/thermo/opt", &spConfig)
		err = calc.RunDFTSinglePoint(optConfig.OrcaPath, "orca")
	}
	if err != nil {
		return fmt.Errorf("error running DFT single point: %w", err)
	}

	// 输出时间差以及当前时间
	utils.FormatDuration(time.Since(start))

	return nil
}
