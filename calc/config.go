package calc

import (
	"bufio"
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"strconv"
	"strings"
)

/*
* config.go
* 该模块主要涉及实现 KYBNMR 运行时所需要读取配置、写入文件功能
*
* @Method:
*
* @Author: Kimariyb
* @Address: XiaMen University
* @Data: 2023-09-21
 */

// Config 结构体，用来存储 ini 文件内的配置。在 KYBNMR 的 ini 文件中，你可以配置以下属性
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
//		preOptArgs(string)：
//		postOptArgs(string):
//		preThreshold(string):
//		postThreshold(string):
//		gauPath(string):
//		orcaPath(string):
//
// [thermo] 根据构象、玻尔兹曼分布计算 NMR 以及偶合参数的配置项
//

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
	PreOptArgs    string
	PostOptArgs   string
	PreThreshold  string
	PostThreshold string
	GauPath       string
	OrcaPath      string
}

type OtherConfig struct {
	Shermopath   string
	MultiwfnPath string
}

// Config 记录 ini 文件配置类
type Config struct {
	DyConfig  DynamicsConfig
	OptConfig OptimizedConfig
	OthConfig OtherConfig
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
	otherConfig := OtherConfig{}

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
	optConfig.PreThreshold = optimizedSection.Key("preThreshold").String()
	optConfig.PostThreshold = optimizedSection.Key("postThreshold").String()
	optConfig.GauPath = optimizedSection.Key("gauPath").String()
	optConfig.OrcaPath = optimizedSection.Key("orcaPath").String()

	// 给 ThermoConfig 赋值
	otherConfig.Shermopath = calculateSection.Key("shermopath").String()
	otherConfig.MultiwfnPath = calculateSection.Key("multiwfnPath").String()

	// 给 config 赋值
	config.DyConfig = dynamicsConfig
	config.OptConfig = optConfig
	config.OthConfig = otherConfig

	return config
}

// ParseXyzFile 用来解析 xyz 文件。将 xyz 中的所有结构都保存在一个 Cluster[] 中
// xyz 文件中的一个结构的第一行为原子数，第二行为能量，第三行到(第三行+原子数-1)行为这个结构的原子坐标
// 接下去就是另外一个结构。我希望把每一个结构都保存在一个 Cluster 中，最后返回这个由 Cluster 组成的 list
// 例如这个 xyz 文件中有两个结构：
// 3
//
//	-44.77460877
//
// C         -2.3118744671        0.7678923498       -1.6678111578
// C         -1.6215849436       -0.3434974558       -1.2274196373
// C         -1.1789998859       -0.4358310737        0.0929450274
// 3
//
//	-44.77460877
//
// C         -2.3118744671        0.7678923498       -1.6678111578
// C         -1.6215849436       -0.3434974558       -1.2274196373
// C         -1.1789998859       -0.4358310737        0.0929450274
func ParseXyzFile(xyzFile string) (ClusterList, error) {
	// 打开XYZ文件
	file, err := os.Open(xyzFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var clusters []Cluster
	var atoms []Atom
	var energy float64
	var lineCount int

	// 逐行读取 XYZ 文件
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		if lineCount == 1 {
			// 解析原子数行
			numAtoms, err := strconv.Atoi(strings.TrimSpace(line))
			if err != nil {
				return nil, fmt.Errorf("invalid atomic number rows：%s", line)
			}

			// 为当前结构创建原子切片
			atoms = make([]Atom, 0, numAtoms)
			energy = 0.0 // 将能量置为空
			continue
		}

		if lineCount == 2 {
			// 解析能量行
			energy, err = strconv.ParseFloat(strings.TrimSpace(line), 64)
			if err != nil {
				// 如果能量行中不是数字而是字符，则将其存为 0.0
				energy = 0.0
			}
			continue
		}

		// 解析原子行
		fields := strings.Fields(line)
		if len(fields) != 4 {
			return nil, fmt.Errorf("invalid atomic rows：%s", line)
		}

		symbol := fields[0]
		x, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return nil, err
		}
		y, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return nil, err
		}
		z, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			return nil, err
		}

		atom := Atom{
			Symbol: symbol,
			X:      x,
			Y:      y,
			Z:      z,
		}

		// 将当前原子添加到原子切片中
		atoms = append(atoms, atom)

		if len(atoms) == cap(atoms) {
			// 当原子数量达到预期时，创建一个新的聚类结构
			cluster := Cluster{
				Atoms:  atoms,
				Energy: energy,
			}

			// 将聚类结构添加到聚类列表中
			clusters = append(clusters, cluster)

			// 重置原子切片和能量，准备下一个结构的解析
			atoms = nil
			energy = 0.0
			lineCount = 0
		}
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return clusters, nil
}

// WriteToXyzFile 向一个标准 xyz 文件中写入信息，同时格式化 xyz 文件
// 如果 xyzFileName 是已经存在的文件，则往文件末尾追加信息
// @param clusters: []Cluster 需要写入的文件信息
// @param xyzFileName: string 需要写入的 xyz 文件的名称
// WriteToXyzFile 向一个标准 XYZ 文件中写入信息，同时格式化 XYZ 文件
// 如果 xyzFileName 是已经存在的文件，则往文件末尾追加信息
// @param clusters: []Cluster 需要写入的文件信息
// @param xyzFileName: string 需要写入的 XYZ 文件的名称
func WriteToXyzFile(clusters ClusterList, xyzFileName string) {
	file, err := os.OpenFile(xyzFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening XYZ file:", err)
		return
	}
	defer file.Close()

	// 写入每个簇的原子数、能量和坐标
	for _, cluster := range clusters {
		// 写入原子数
		_, err = file.WriteString(fmt.Sprintf("  %d\n", len(cluster.Atoms)))
		if err != nil {
			fmt.Println("Error writing atom count to XYZ file:", err)
			return
		}
		// 写入能量
		_, err = file.WriteString(fmt.Sprintf("\t\t%.8f\n", cluster.Energy))
		if err != nil {
			fmt.Println("Error writing energy to XYZ file:", err)
			return
		}

		// 写入每个原子的坐标
		for _, atom := range cluster.Atoms {
			_, err = file.WriteString(fmt.Sprintf("%2s \t\t%14.10f \t\t%14.10f \t\t%14.10f\n", atom.Symbol, atom.X, atom.Y, atom.Z))
			if err != nil {
				fmt.Println("Error writing atom coordinates to XYZ file:", err)
				return
			}
		}
	}

	fmt.Println("Hint: XYZ file written successfully.")
}
