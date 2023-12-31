package main

import (
	"kybnmr/run"
)

/*
* kybnmr.go
* 这是 KYBNMR 程序的主入口，完成主函数逻辑，以及调用其他编写好的包。
* 思路: 1. 首先调用 xtb 在 GFN0-xTB 做动力学任务，让输入的目标分子能够在体系中足够的变化，产生足够多的轨迹
*      2. 接着再调用 xtb/crest 再用 GFN2-xTB 下做预优化，同时考虑溶剂模型
*      3. 接着调用 Gaussian 在真空下做优化和振动分析得到较可靠的结构和自由能热校正量，
* 		  再调用 ORCA 对每个结构在高级别下结合 SMD 模型表现的水环境下算高精度单点能，
*		  二者相加得到水环境下的高精度自由能
* [每一步都需要经过 KYBNMR 的 Double Check 检测，结构集是否合格]
*
* @Author: Kimariyb
* @Address: XiaMen University
* @Data: 2023-09-20
 */

func main() {
	// KYBNMR 主程序运行
	KYBNMR := run.NewKYBNMR()
	KYBNMR.ParseArgsToRun()
}
