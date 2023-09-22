package calc

import (
	"fmt"
	"os/exec"
)

/*
* execute.go
* 1. 该模块用来调用 xtb 做分子动力学模拟，或者调用 xtb 做半经验优化。
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
		fmt.Println("xtb has been successfully detected.")
		return true
	} else {
		// 如果调用失败，则打印 xtb is not detected, please install xtb. 同时返回 False
		fmt.Println("xtb is not detected, please install xtb.")
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

func XtbExecuteMD() {

}

func XtbExecuteOpt() {

}
