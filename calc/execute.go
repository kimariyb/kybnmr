package calc

import (
	"fmt"
	"io"
	"io/ioutil"
	"kybnmr/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

/*
* execute.go
* 1. 该模块用来调用 xtb 做分子动力学模拟，或者调用 crest 做半经验优化。
* 2. 该模块用来调用 Gaussian 和 Orca 做优化和能量计算
*
* @Version:
* 	xtb: 6.6.0 (8843059)
* 	Gaussian: A.03/C.01
*	Orca: 5.0.4
*
* @Author: Kimariyb
* @Address: XiaMen University
* @Data: 2023-09-21
 */

// IsExistXtb 检查环境变量中是否存在 Xtb 程序。
// 返回一个布尔值，表示是否存在 Xtb 程序。
func IsExistXtb() bool {
	// 在命令行调用 xtb --version 命令，如果调用成功，就说明存在 Xtb 程序
	cmd := exec.Command("xtb", "--version")
	err := cmd.Run()
	if err == nil {
		// 如果调用成功，则打印 xtb has been successfully detected. 同时返回 True.
		fmt.Println("Hint: xtb has been successfully detected.")
		return true
	} else {
		// 如果调用失败，则打印 xtb is not detected, please install xtb. 同时返回 False
		fmt.Println("Error: xtb is not detected, please install xtb.")
		return false
	}
}

// XtbExecuteMD 调用 xtb 程序执行分子动力学模拟
// @param: dyConfig(DynamicsConfig)
// @param: xybFile(string)
// dy.inp 模板为
// $md
//
//	temp=${dyConfig.temperature} # in K
//	time=${dyConfig.time}  # in ps
//	dump=${dyConfig.dump}  # in fs
//	step=${dyConfig.step}  # in fs
//	velo=${dyConfig.velo}
//	nvt =${dyConfig.nvt}
//	hmass=${dyConfig.hmass}
//	shake=${dyConfig.shake}
//	sccacc=${dyConfig.sccacc}
//
// $end
func XtbExecuteMD(dyConfig *DynamicsConfig, xyzFile string) error {
	// 检查 temp 文件夹是否存在
	_, err := os.Stat("temp")
	if os.IsNotExist(err) {
		// 如果 temp 文件夹不存在，则创建它
		err = os.Mkdir("temp", 0755)
		if err != nil {
			fmt.Println("Error creating temp directory:", err)
			return nil
		}
	}

	// 创建一个临时模板文件 md.inp
	templateText := `$md
	temp={{.Temperature}}
	time={{.Time}}
	dump={{.Dump}}
	step={{.Step}}
	velo={{.Velo}}
	nvt={{.Nvt}}
	hmass={{.Hmass}}
	shake={{.Shake}}
	sccacc={{.Sccacc}}
$end
`
	// 在当前运行目录下的 temp 文件夹中，创建一个临时文件 md.inp
	// 如果没有 temp 文件，则新建一个 temp 文件夹
	tempFile, err := os.Create(filepath.Join("temp", "md.inp"))
	if err != nil {
		fmt.Println("Error creating temp file:", err)
		return nil
	}
	// 最后关闭并删除 md.inp 文件
	defer func() {
		err := tempFile.Close()
		if err != nil {
			return
		}
		err = os.Remove(tempFile.Name())
		if err != nil {
			return
		}
	}()

	// 将 dyConfig 中的数据写入 dy.inp 中
	tmpl := template.Must(template.New("md.inp").Parse(templateText))
	err = tmpl.Execute(tempFile, dyConfig)
	if err != nil {
		fmt.Println("Error writing template to file", err)
		return nil
	}

	// 执行 xtb 程序
	// 首先，检测当前环境中是否存在 xtb 程序
	if IsExistXtb() {
		// 如果存在，则继续执行
		// 构建 xtb 命令行参数
		otherArgs := utils.SplitStringBySpace(dyConfig.DynamicsArgs)
		cmdArgs := []string{xyzFile, "--input", tempFile.Name(), dyConfig.DynamicsArgs}
		cmdArgs = append(cmdArgs, otherArgs...)
		//创建 xtb 命令对象
		cmd := exec.Command("xtb", cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		//执行 xtb 命令，并且在命令行中显示 xtb 运行的输出
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error executing xtb:", err)
			return nil
		}

		// 成功结束后，打印信息
		fmt.Println("xtb MD simulation completed successfully.")

		// 将 xtb 生成的文件全部移动到 temp 文件夹中
		utils.MoveAllFileButKeepFile([]string{"KYBNMR", "kybnmr", xyzFile, "*.ini", "xtb.trj", "GauTemplate.gjf", "OrcaTemplate.inp"}, "temp")
		// 将生成的 xtb.trj 文件修改为 dynamic.xyz
		utils.RenameFile("xtb.trj", "dynamics.xyz")
	}

	return nil
}

// RunCrestOptimization 调用 crest 程序并行执行 xtb 方法
func RunCrestOptimization(args string, inputFile string, outputFile string, finalFile string) {
	// 拿到 bin 目录下的 crest 程序的路径，并直接调整为绝对路径
	crestPath, err := filepath.Abs(filepath.Join("bin", "crest"))
	if err != nil {
		fmt.Println("Error getting crest program path:", err)
		return
	}

	// 根据 optConfig 配置中的内容，调用 crest 进行优化
	otherArgs := utils.SplitStringBySpace(args)
	cmdArgs := []string{"--mdopt", inputFile}
	cmdArgs = append(cmdArgs, otherArgs...)

	// 创建 crest 命令对象
	cmd := exec.Command(crestPath, cmdArgs...)
	// 设置标准输出和标准错误输出
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 执行 crest 命令，如果运行 crest 报错，则直接退出，如果没有报错，则继续
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error executing crest:", err)
		return
	} else {
		fmt.Println("Crest optimization completed successfully.")
		// 必须跳过的文件
		SkipFileName := []string{"KYBNMR", "kybnmr", "*.ini", "xtb.trj", inputFile, "GauTemplate.gjf", "OrcaTemplate.inp", "*.out", "*.xyz"}
		// 将 crest 生成的文件全部移动到 temp 文件夹中
		utils.MoveAllFileButKeepFile(SkipFileName, "temp")
		// 将 crest_ensemble.xyz 文件修改为指定的输出文件名
		utils.RenameFile(outputFile, finalFile)
	}
}

// XtbExecutePreOpt 调用 Xtb 对体系做预优化，由于 xtb 不支持并行，因此这里直接使用 xtb 升级版 crest
// crest 已经在本程序的 bin 目录下了，并不需要手动下载
func XtbExecutePreOpt(optConfig *OptimizedConfig, xyzFile string) {
	RunCrestOptimization(optConfig.PreOptArgs, xyzFile, "crest_ensemble.xyz", "pre_opt.xyz")
}

// XtbExecutePostOpt 调用 xtb 对体系进行进一步优化
func XtbExecutePostOpt(optConfig *OptimizedConfig, xyzFile string) {
	RunCrestOptimization(optConfig.PostOptArgs, xyzFile, "crest_ensemble.xyz", "post_opt.xyz")
}

// RunDFTOptimization 调用指定的软件对当前文件下的 gjf 文件进行优化运算
// 运算的原理：首先获取运行目录下的 GauTemplate.gjf，这是一个 Gaussian 输入文件的模板文件
// 将文件中的 [GEOMETRY] 用实际的原子坐标替换后，在 thermo/opt 文件夹中生成一个新的 Gaussian gjf 输入文件
// 接着调用 Gaussian 运行这个 gjf 输入文件后，直接在 thermo/opt 文件夹中生成 out 文件
// Clusters 每有一个 Cluster 就按照上述方法运行一次 Gaussian，直到 Clusters 中的所有元素都被遍历完。
// # opt freq b3lyp/6-31g* int=fine scrf(solvent=CHCl3)
//
// # Template file
//
// 0 1
// [GEOMETRY]
func RunDFTOptimization(softwarePath string, templateFile string, clusters ClusterList, softwareName string) error {
	// 读取模板文件内容
	templateContent, err := ioutil.ReadFile(templateFile)
	if err != nil {
		fmt.Println("Error reading template file:", err)
		return nil
	}

	// 创建 thermo/opt 文件夹（如果不存在）
	optFolderPath := "thermo/opt"
	err = os.MkdirAll(optFolderPath, 0755)
	if err != nil {
		fmt.Println("Error creating opt folder:", err)
		return nil
	}

	for i, cluster := range clusters {
		// 生成新的输入文件名
		inputFileName := fmt.Sprintf("cluster-opt%d%s", i+1, filepath.Ext(templateFile))
		// 生成新的输出文件名
		outFileName := fmt.Sprintf("cluster-opt%d.out", i+1)
		inputFilePath := filepath.Join(optFolderPath, inputFileName)

		// 替换模板文件中的 [GEOMETRY] 标记
		inputContent := strings.Replace(string(templateContent), "[GEOMETRY]", cluster.ToXYZString(), 1)
		// 追加两行空格
		inputContent += "\n\n"

		// 将新的输入文件写入磁盘
		// 请注意，一定要在末尾追加两行空格
		err = ioutil.WriteFile(inputFilePath, []byte(inputContent), 0644)
		if err != nil {
			fmt.Println("Error writing input file:", err)
			return nil
		}

		var cmd *exec.Cmd

		// 调用指定的软件运行输入文件
		if strings.EqualFold(softwareName, "Gaussian") {
			cmd = exec.Command("bash", "-c", fmt.Sprintf("%s < %s > %s", softwarePath, inputFilePath, outFileName))
		} else if strings.EqualFold(softwareName, "Orca") {
			cmd = exec.Command("bash", "-c", fmt.Sprintf("%s %s > %s", softwarePath, inputFilePath, outFileName))
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// 输出正在运行 xxx.gjf 或者 xxx.inp
		fmt.Printf("Hint: %s is Running: %s\n", softwareName, inputFileName)

		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error executing %s: %s\n", softwareName, err)
			return nil
		}

		fmt.Printf("Hint: %s calculation completed for cluster %d\n", softwareName, i+1)
	}
	fmt.Println()
	fmt.Printf("Hint: %s optimization completed successfully.\n", softwareName)

	// 将所有生成的 out 文件都放进 ./thermo/opt 中
	utils.MoveFileForType(".out", "thermo/opt")

	return nil
}

// RunDFTSinglePoint 调用 DFT 程序计算单点能
func RunDFTSinglePoint(softwarePath string, softwareName string) error {
	// 扫描 thermo/sp 文件夹下的所有 .inp 或 .gjf 文件
	// 如果 softwareName 为 Gaussian 就是 .gjf，如果是 Orca 就是 .inp
	inputFiles, err := utils.ScanInputFiles("thermo/sp", softwareName)

	for i, inputFile := range inputFiles {
		// 生成新的输出文件名
		outFileName := fmt.Sprintf("cluster-sp%d.out", i+1)

		var cmd *exec.Cmd

		// 调用指定的软件运行输入文件
		if strings.EqualFold(softwareName, "Gaussian") {
			cmd = exec.Command("bash", "-c", fmt.Sprintf("%s < %s > %s", softwarePath, inputFile, outFileName))
		} else if strings.EqualFold(softwareName, "Orca") {
			cmd = exec.Command("bash", "-c", fmt.Sprintf("%s %s > %s", softwarePath, inputFile, outFileName))
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// 输出正在运行 xxx.gjf 或者 xxx.inp
		fmt.Printf("Hint: %s is Running: %s\n", softwareName, inputFile)

		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error executing %s: %s\n", softwareName, err)
			return nil
		}

		fmt.Printf("Hint: %s calculation completed for cluster %d\n", softwareName, i+1)

	}

	fmt.Println()
	fmt.Printf("Hint: %s single point energy completed successfully.\n", softwareName)

	// 将所有生成的 out 文件都放进 ./thermo/sp 中
	utils.MoveFileForType(".out", "thermo/sp")

	return nil

}

// BatchMTFToGenerateFile 批量通过 Multiwfn 产生文件
// 首先扫描指定文件夹下的所有的 out 文件
// 接着批量调用 GenerateFileFromMTF() 生成 inp、gjf 文件
// 将当前文件夹下的所有 inp 文件全部都转移到 ./thermo/sp 文件中
func BatchMTFToGenerateFile(softwareName string, targetFolder string, spConfig *SingleConfig) error {
	// 获取指定文件夹下的所有 out 文件
	files, err := ioutil.ReadDir(targetFolder)
	if err != nil {
		return fmt.Errorf("unable to read target folder: %v", err)
	}

	// 创建 thermo/sp 文件夹（如果不存在）
	spFolderPath := "thermo/sp"
	err = os.MkdirAll(spFolderPath, 0755)
	if err != nil {
		fmt.Println("Error creating opt folder:", err)
		return nil
	}

	// 批量生成文件并移动
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".out" {
			// 生成 inp 文件
			inpFile := GenerateFileFromMTF(softwareName, filepath.Join(targetFolder, file.Name()), spConfig)

			// 移动 inp 文件到指定文件夹
			err := utils.MoveFile(inpFile, "/thermo/sp")
			if err != nil {
				fmt.Printf("Unable to move files %s to the destination folder: %v\n", inpFile, err)
			}
		}
	}

	return nil
}

// GenerateFileFromMTF
// 使用 Multiwfn 处理 out 文件得到 orca 的 inp 文件或 gaussian 的 gjf 文件
// 参数 targetFileType 需要转化的目标文件，例如 .inp; .gjf
// 参数 sourceFile 为需要转化的 out 文件
// 参数 spConfig，SingleConfig 的一个实例
func GenerateFileFromMTF(softwareName string, sourceFile string, spConfig *SingleConfig) string {
	// 创建一个命令对象，执行 Multiwfn 程序
	cmd := exec.Command(spConfig.MultiwfnPath, sourceFile)

	// 获取标准输入管道
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	// 启动命令
	err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting process:", err)
		return ""
	}

	// 取到 sourceFile 文件的文件名，同时拼接上 targetFileType
	// 例如 sourceFile 为 hello.out，targetFileType 为 .inp
	// 拼接后为 hello.inp
	outputName := ""

	// 在 spConfig.MultiwfnCommand 前追加一些内容
	if softwareName == "orca" {
		// 如果 spCalcName 选择的是 orca 则使用 orca 命令赋值
		outputName = utils.ConcatenateFileName(sourceFile, ".inp")
		spConfig.MultiwfnCommand = utils.PrependToSlice(spConfig.MultiwfnCommand, "100", "2", "12", outputName)
	} else if softwareName == "gaussian" {
		// 如果 spCalcName 选择的是 gaussian 则使用 gaussian 命令赋值
		outputName = utils.ConcatenateFileName(sourceFile, ".gjf")
		spConfig.MultiwfnCommand = utils.PrependToSlice(spConfig.MultiwfnCommand, "100", "2", "10", outputName)
	}

	// 向命令的标准输入写入指令
	for _, line := range spConfig.MultiwfnCommand {
		_, err := io.WriteString(stdin, line+"\n")
		if err != nil {
			fmt.Println("Error:", err)
			return ""
		}
	}

	// 关闭标准输入管道
	stdin.Close()

	// 等待命令执行完成
	err = cmd.Wait()

	return outputName
}
