package calc

import (
	"CalcNMR/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

/*
* execute.go
* 1. 该模块用来调用 xtb 做分子动力学模拟，或者调用 crest 做半经验优化。
* 2. 该模块用来调用 Gaussian 和 Orca 做优化和能量计算
*
* @Method:
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

// IsExistGaussian 检查环境变量中否存在 Gaussian 程序
// @Param: gauPath(string): Gaussian 可执行文件的路径
// @Return: bool
func IsExistGaussian(gauPath string) {
	// 根据 Gaussian 可执行文件的路径，确定 Gaussian 的执行命令
}

// IsExistOrca 检查环境变量中否存在 Orca 程序
// @Param: orcaPath(string): Orca 可执行文件的路径
// @Return: bool
func IsExistOrca(orcaPath string) {
	// 根据 Orca 可执行文件的路径，确定 Orca 的执行命令

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
func XtbExecuteMD(dyConfig *DynamicsConfig, xyzFile string) {
	// 检查 temp 文件夹是否存在
	_, err := os.Stat("temp")
	if os.IsNotExist(err) {
		// 如果 temp 文件夹不存在，则创建它
		err = os.Mkdir("temp", 0755)
		if err != nil {
			fmt.Println("Error creating temp directory:", err)
			return
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
		return
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
		return
	}

	// 执行 xtb 程序
	// 首先，检测当前环境中是否存在 xtb 程序
	if IsExistXtb() {
		// 如果存在，则继续执行
		// 构建 xtb 命令行参数
		cmdArgs := []string{xyzFile, "--input", tempFile.Name(), "--omd", "--gfn", "0"}

		//创建 xtb 命令对象
		cmd := exec.Command("xtb", cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		//执行 xtb 命令，并且在命令行中显示 xtb 运行的输出
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error executing xtb:", err)
			return
		}
		// 成功结束后，打印信息
		fmt.Println("xtb MD simulation completed successfully.")
	}

	// 将 xtb 生成的文件全部移动到 temp 文件夹中
	utils.RemoveTempFolder([]string{"CalcNMR", xyzFile, "*.ini", "xtb.trj"})
	// 将生成的 xtb.trj 文件修改为 dynamic.xyz
	utils.RenameFile("xtb.trj", "dynamic.xyz")
}

// XtbExecuteOpt 调用 Xtb 对体系做预优化，由于 xtb 不支持并行，因此这里直接使用 xtb 升级版 crest
// crest 已经在本程序的 bin 目录下了，并不需要手动下载
func XtbExecuteOpt(optConfig *OptimizedConfig, xyzFile string) {
	// 拿到 bin 目录下的 crest 程序的路径，并直接调整为绝对路径
	crestPath, err := filepath.Abs(filepath.Join("bin", "crest"))
	if err != nil {
		fmt.Println("Error getting crest program path:", err)
		return
	}

	// 根据 optConfig 配置中的内容，调用 crest 做预优化
	cmdArgs := []string{
		"-mdopt", "dynamic.xyz", optConfig.PreOptArgs,
	}

	// 创建 crest 命令对象
	cmd := exec.Command(crestPath, cmdArgs...)
	// 设置标准输出和标准错误输出
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 执行crest命令
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error executing crest:", err)
		return
	}

	fmt.Println("crest optimization completed successfully.")

	// 将 crest 生成的文件全部移动到 temp 文件夹中
	utils.RemoveTempFolder([]string{"dynamic.xyz", "CalcNMR", "*.ini", xyzFile, "xtb.trj", "crest_ensemble.xyz"})
	// 将 crest_ensemble.xyz 文件修改为 pre_opt.xyz
	utils.RenameFile("crest_ensemble.xyz", "pre_opt.xyz")
}
