package calc

import (
	"fmt"
	"io"
	"io/ioutil"
	"kybnmr/utils"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// ReadClusterListFromOut 扫描指定文件夹下的所有的 out 文件，
// 调用 ParseOutFile 方法读取所有 out 文件，并且返回成 ClusterList
// 传入的参数：
//   - softwareName string: 使用的程序
func ReadClusterListFromOut(softwareName string) (ClusterList, error) {
	var clusterList ClusterList

	// 获取主程序运行文件夹的绝对路径
	currentDir, err := os.Getwd()
	if err != nil {
		return clusterList, err
	}

	// 构建 thermo/opt 文件夹的完整路径
	targetFolder := filepath.Join(currentDir, "thermo/opt")
	// 首先扫描指定文件夹下的所有 out 文件
	files, err := ioutil.ReadDir(targetFolder)
	if err != nil {
		return clusterList, err
	}
	// 遍历文件夹中的每个文件
	for _, file := range files {
		// 检查文件是否为 out 文件
		if strings.HasSuffix(file.Name(), ".out") {
			// 获取 out 文件的完整路径
			outFilePath := filepath.Join(targetFolder, file.Name())
			// 解析 out 文件并将聚类添加到聚类列表中
			cluster, err := ParseOutFile(softwareName, outFilePath)
			if err != nil {
				return clusterList, err
			}
			clusterList = append(clusterList, cluster)
		}
	}

	return clusterList, nil
}

// RunDFTSinglePoint 调用 DFT 程序进行单点任务
// 运算的原理：首先获取运行目录下的 OrcaTemplate.gjf，这是一个 Orca 输入文件的模板文件
// 将文件中的 [GEOMETRY] 用实际的原子坐标替换后，在 thermo/sp 文件夹中生成一个新的 Orca inp 输入文件
// 接着调用 Orca 运行这个 inp 输入文件后，直接在 thermo/sp 文件夹中生成 out 文件
// Clusters 每有一个 Cluster 就按照上述方法运行一次 Orca，直到 Clusters 中的所有元素都被遍历完。
func RunDFTSinglePoint(softwarePath string, templateFile string, clusters ClusterList, softwareName string) error {
	// 读取模板文件内容
	templateContent, err := ioutil.ReadFile(templateFile)
	if err != nil {
		fmt.Println("Error reading template file:", err)
		return nil
	}

	// 创建 thermo/sp 文件夹（如果不存在）
	optFolderPath := "thermo/sp"
	err = os.MkdirAll(optFolderPath, 0755)
	if err != nil {
		fmt.Println("Error creating opt folder:", err)
		return nil
	}

	for i, cluster := range clusters {
		// 生成新的输入文件名
		inputFileName := fmt.Sprintf("cluster-sp%d%s", i+1, filepath.Ext(templateFile))
		// 生成新的输出文件名
		outFileName := fmt.Sprintf("cluster-sp%d.out", i+1)
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
	fmt.Printf("Hint: %s single point energy completed successfully.\n", softwareName)

	// 将所有生成的 out 文件都放进 ./thermo/sp 中
	utils.MoveFileForType(".out", "thermo/sp")

	return nil
}

// RunShermoToBolzmann 调用 Shermo 计算 Bolzmann 分布
// 首先定位到当前程序运行的 thermo/opt 文件夹下，在 thermo/opt 新建一个 txt 文件，文件模板内容如下：
// [FileName] [Energy]
// ................
// FileName 和 Energy 都是 resultCollection 中每一个 result 结构体的属性
// 接着调用 shermo 执行这个 txt 文件，同时在屏幕上显示输出的内容
// 参数：
//   - resultCollection: []ShermoResult: ShermoResult 组成的 slice
//   - ShermoResult: FileName string: 文件的路径
//   - ShermoResult: Energy   string: 能量
//   - shermoPath: string shermo 程序的运行路径
func RunShermoToBolzmann(resultCollection []ShermoResult, shermoPath string) error {
	// 获取主程序运行文件夹的绝对路径
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// 获取 currentDir/thermo/opt 下的所有 out 文件的名字
	files, err := ioutil.ReadDir(filepath.Join(currentDir, "thermo/opt"))
	if err != nil {
		return err
	}

	// 新建一个 filesNames 接收 files 的 Name
	var filesNames []string
	for _, file := range files {
		filesNames = append(filesNames, file.Name())
	}

	// 创建 txt 文件并写入内容
	txtFilePath := filepath.Join(currentDir, "thermo/opt/shermo.txt")
	if err := createInputFile(txtFilePath, filesNames, resultCollection); err != nil {
		return err
	}

	// 创建输出文件
	outputFile := filepath.Join(currentDir, "thermo/opt/output.txt")

	// 通过命令行运行 Shermo
	cmd := exec.Command(shermoPath, txtFilePath)
	result, err := cmd.CombinedOutput()
	if err == nil {
		fmt.Println()
		fmt.Printf("Hint: Shermo completed successfully on file %s.\n\n", txtFilePath)
		contents := string(result)
		// 写入输出数据到文件
		outputFile := filepath.Join(outputFile + ".txt")
		err = ioutil.WriteFile(outputFile, []byte(contents), 0644)
		if err != nil {
			fmt.Printf("Error writing output file: %v\n", err)
		}
	} else {
		fmt.Println()
		fmt.Printf("Hint: Shermo execution failed on file %s.\n", txtFilePath)
	}

	return nil
}

func createInputFile(filePath string, optFilePaths []string, resultCollection []ShermoResult) error {
	var lines []string
	for i, result := range resultCollection {
		line := fmt.Sprintf("%s %s", optFilePaths[i], result.Energy)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	err := ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return err
	}

	return nil
}

func FindLastMatch(contents string, regex *regexp.Regexp, groupIndex int) (string, error) {
	// 使用正则表达式在字符串中查找所有匹配项
	matches := regex.FindAllStringSubmatch(contents, -1)
	// 获取第二个匹配项
	if len(matches) >= 2 {
		secondMatch := matches[1]
		if len(secondMatch) > groupIndex {
			return secondMatch[groupIndex], nil
		}
	}
	return "", fmt.Errorf("no energy found")
}

func GetGaussianEnergy() []ShermoResult {
	// 创建一个切片用来存放每一个文件对应的 results
	var resultsCollection []ShermoResult

	// 获取主程序运行文件夹的绝对路径
	currentDir, err := os.Getwd()
	if err != nil {
		return resultsCollection
	}

	// 构建 thermo/sp 文件夹的完整路径
	targetFolder := filepath.Join(currentDir, "thermo/sp")

	// 读取目标文件夹下的所有文件
	files, err := os.ReadDir(targetFolder)
	if err != nil {
		fmt.Println("Error: failed to read directory", err)
		return resultsCollection
	}

	// 遍历文件列表并处理每个文件
	for _, file := range files {
		// 获取文件的完整路径
		filePath := filepath.Join(targetFolder, file.Name())

		// 通过 filePath 打开文件
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("Error: Unable to open the file", err)
			continue
		}
		defer file.Close()

		// 读取文件内容为 Bytes
		contentsBytes, err := io.ReadAll(file)
		// Bytes 转化为字符串
		contentsString := string(contentsBytes)
		if err != nil {
			fmt.Println("Error: Failed to read", err)
			continue
		}

		// 替换空格
		re := regexp.MustCompile(`\s+`)
		contentsString = re.ReplaceAllString(contentsString, "")

		if err != nil {
			fmt.Println("Error: Failed to read", err)
			continue
		}

		// 使用正则表达式搜索 gaussian 单点能
		ccsdTRegex := regexp.MustCompile(`CCSD\(T\)=\s*(-?\d+\.\d+)`)
		mp2Regex := regexp.MustCompile(`MP2=\s*(-?\d+\.\d+)`)
		hfRegex := regexp.MustCompile(`HF=\s*(-?\d+\.\d+)`)

		// 首先匹配是否存在 CCSD(T) 的能量，如果存在则直接读取，并将结果保存在 results 中
		ccsdTEnergy, err := FindLastMatch(contentsString, ccsdTRegex, 1)
		if err == nil {
			fileResults := ShermoResult{
				FileName: filePath,
				Energy:   ccsdTEnergy,
			}
			fmt.Println("The Single Point Energy [CCSD(T)] of " + file.Name() + " is : " + ccsdTEnergy)

			resultsCollection = append(resultsCollection, fileResults)
			continue
		}
		// 如果不存在 CCSD(T) 的能量，但是存在 MP2 能量，则将 MP2 结果保存在 results 中
		mp2Energy, err := FindLastMatch(contentsString, mp2Regex, 1)
		if err == nil {
			fileResults := ShermoResult{
				FileName: filePath,
				Energy:   mp2Energy,
			}
			fmt.Println("The Single Point Energy [MP2] of " + file.Name() + " is : " + mp2Energy)

			resultsCollection = append(resultsCollection, fileResults)
			continue
		}
		// 如果不存在 CCSD(T) 和 MP2 的能量，但是存在 HF 能量，则将 HF 结果保存在 results 中
		hfEnergy, err := FindLastMatch(contentsString, hfRegex, 1)
		if err == nil {
			fileResults := ShermoResult{
				FileName: filePath,
				Energy:   hfEnergy,
			}
			fmt.Println("The Single Point Energy [HF] of " + file.Name() + " is : " + mp2Energy)
			resultsCollection = append(resultsCollection, fileResults)
			continue
		}

		fmt.Println("No energy found", filePath)
	}

	return resultsCollection
}

func GetOrcaEnergy() []ShermoResult {
	// 创建一个切片用来存放每一个文件对应的 results
	var resultsCollection []ShermoResult

	// 获取主程序运行文件夹的绝对路径
	currentDir, err := os.Getwd()
	if err != nil {
		return resultsCollection
	}

	// 构建 thermo/sp 文件夹的完整路径
	targetFolder := filepath.Join(currentDir, "thermo/sp")

	// 读取目标文件夹下的所有文件
	files, err := os.ReadDir(targetFolder)
	if err != nil {
		fmt.Println("Error: failed to read directory", err)
		return resultsCollection
	}

	// 遍历文件列表并处理每个文件
	for _, file := range files {
		// 获取文件的完整路径
		filePath := filepath.Join(targetFolder, file.Name())

		// 通过 filePath 打开文件
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("Error: Unable to open the file", err)
			continue
		}
		defer file.Close()

		// 读取文件内容为 Bytes
		contentsBytes, err := io.ReadAll(file)
		// Bytes 转化为字符串
		contentsString := string(contentsBytes)
		if err != nil {
			fmt.Println("Error: Failed to read", err)
			continue
		}

		// 使用正则表达式搜索 orca 单点能
		energyRegex := regexp.MustCompile(`FINAL SINGLE POINT ENERGY\s+(-?\d+\.\d+)`)

		// 查找匹配的能量值
		matches := energyRegex.FindAllStringSubmatch(contentsString, -1)
		if len(matches) > 0 {
			// 查找文件中最后一个匹配项的能量值
			energy := matches[len(matches)-1][1]
			fmt.Println("Energy:", energy)

			// 创建 results 结构体对象
			fileResults := ShermoResult{
				FileName: filePath,
				Energy:   energy,
			}

			resultsCollection = append(resultsCollection, fileResults)
		} else {
			fmt.Println("No energy found", filePath)
		}
	}

	return resultsCollection
}
