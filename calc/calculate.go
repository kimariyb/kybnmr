package calc

import (
	"gopkg.in/ini.v1"
)

/*
* calculate.go
*
* @Method:
*
* @Author: Kimariyb
* @Address: XiaMen University
* @Data: 2023-09-21
 */

// Config 结构体，用来存储 ini 文件内的配置。在 CalcNMR 的 ini 文件中，你可以配置以下属性
// [dynamics] 使用 xtb 做动力学的配置项
//		temperature(float): 温度，单位为 K
//		time(float): 时间，单位为 ps
//		dump(float): 每间隔多少时间往轨迹文件里写入一次，单位为 fs
//		step(float): 步长，单位为 fs
//		velo(bool): 是否输出速度
//		nvt(bool): 是否开启 NVT 模拟
//		hmass(int): 氢原子的质量是实际的多少倍
//		shake(int): 将与氢有关的化学键距离都用 SHAKE 算法约束住
//		sccacc(float): 动态 xTB 计算的准确性
//		dynamicArgs(string): 运行 xtb 做动力学的命令
//
// [optimized] 使用 xtb 做预优化的配置项、使用 Gaussian 和 orca 做进一步优化的配置项
//		preOptArgs：
//		postOptArgs:
//
// [calculate] 根据构象、玻尔兹曼分布计算 NMR 以及偶合参数的配置项

// DynamicsConfig ini 文件中动力学部分的配置文件
type DynamicsConfig struct {
	Temperature  float64
	Time         float64
	Dump         float64
	Step         float64
	Velo         bool
	Nvt          bool
	Hmass        int
	Shake        int
	Sccacc       float64
	DynamicsArgs string
}

// OptimizedConfig ini 文件中优化部分的配置文件
type OptimizedConfig struct {
	PreOptArgs  string
	PostOptArgs string
	Gaupath     string
	Orcapath    string
}

type CalculateConfig struct {
	Shermopath string
}

// Config 记录 ini 文件配置类
type Config struct {
	DyConfig   DynamicsConfig
	OptConfig  OptimizedConfig
	CalcConfig CalculateConfig
}

// ParseConfigFile 解析符合条件的 ini 文件，并且返回一个 Config 对象
func ParseConfigFile(configFile string) *Config {
	// 声明一个 Config 结构体
	config := &Config{}
	// 解析 ini 文件
	iniFile, err := ini.Load(configFile)
	if err != nil {
		return nil
	}
	// 分别解析 ini 文件中的 [dynamics]、[optimized]、[calculate] 组分别存储在
	// DynamicsConfig、OptimizedConfig、CalculateConfig 结构体中
	// 最后将 DynamicsConfig、OptimizedConfig、CalculateConfig 结构体存储在 Config 中
	dynamicsSection := iniFile.Section("dynamics")
	optimizedSection := iniFile.Section("optimized")
	calculateSection := iniFile.Section("calculate")

	// 声明一个 dynamicsConfig、OptimizedConfig、CalclateConfig
	dynamicsConfig := DynamicsConfig{}
	optConfig := OptimizedConfig{}
	calcConfig := CalculateConfig{}

	// 给 dynamicsConfig 赋值
	dynamicsConfig.Temperature, _ = dynamicsSection.Key("temperature").Float64()
	dynamicsConfig.Time, _ = dynamicsSection.Key("time").Float64()
	dynamicsConfig.Step, _ = dynamicsSection.Key("step").Float64()
	dynamicsConfig.Dump, _ = dynamicsSection.Key("dump").Float64()
	dynamicsConfig.Nvt, _ = dynamicsSection.Key("nvt").Bool()
	dynamicsConfig.Velo, _ = dynamicsSection.Key("velo").Bool()
	dynamicsConfig.Shake, _ = dynamicsSection.Key("shake").Int()
	dynamicsConfig.Hmass, _ = dynamicsSection.Key("hmass").Int()
	dynamicsConfig.Sccacc, _ = dynamicsSection.Key("sccacc").Float64()
	dynamicsConfig.DynamicsArgs = dynamicsSection.Key("dynamicsArgs").String()

	// 给 optConfig 赋值
	optConfig.PreOptArgs = optimizedSection.Key("preOptArgs").String()
	optConfig.PostOptArgs = optimizedSection.Key("postOptArgs").String()
	optConfig.Gaupath = optimizedSection.Key("gaupath").String()
	optConfig.Orcapath = optimizedSection.Key("orcapath").String()

	// 给 calcConfig 赋值
	calcConfig.Shermopath = calculateSection.Key("shermopath").String()

	// 给 config 赋值
	config.DyConfig = dynamicsConfig
	config.OptConfig = optConfig
	config.CalcConfig = calcConfig

	return config
}

// DoubleCheck 用于 CalcNMR 检查构象是否合理，以及是否存在重复结构
// DoubleCheck 函数是整个 CalcNMR 最核心的步骤
// @param: thresholds(float): 查找的阈值
func DoubleCheck(thresholds float64) {

}
