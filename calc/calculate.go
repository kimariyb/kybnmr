package calc

/*
* calculate.go
*
* @Method:
*
* @Author: Kimariyb
* @Address: XiaMen University
* @Data: 2023-09-21
 */

// TomlConfig 结构体，用来存储 toml 文件内的配置。在 CalcNMR 的 toml 文件中，你可以配置以下属性
// [dynamics] 使用 xtb 做动力学的配置项
//
// [pre-optimized] 使用 xtb 做预优化的配置项
//
// [post-optimized] 使用 gaussian 以及 orca 做进一步优化的配置项
//
// [calculate] 根据构象、玻尔兹曼分布计算 NMR 以及偶合参数的配置项
type TomlConfig struct {
}
