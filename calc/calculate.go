package calc

import (
	"bufio"
	"fmt"
	"gopkg.in/ini.v1"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
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

// Atom 记录每一个原子的结构体
type Atom struct {
	Symbol  string
	X, Y, Z float64
}

// Cluster xyz 文件中所记录的结构和能量
type Cluster struct {
	Atoms  []Atom
	Energy float64
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
func ParseXyzFile(xyzFile string) ([]Cluster, error) {
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

	// 逐行读取XYZ文件
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		if lineCount == 1 {
			// 解析原子数行
			numAtoms, err := strconv.Atoi(strings.TrimSpace(line))
			if err != nil {
				return nil, fmt.Errorf("无效的原子数行：%s", line)
			}

			// 为当前结构创建原子切片
			atoms = make([]Atom, 0, numAtoms)
			energy = 0.0
			continue
		}

		if lineCount == 2 {
			// 解析能量行
			energy, err = strconv.ParseFloat(strings.TrimSpace(line), 64)
			if err != nil {
				return nil, fmt.Errorf("无效的能量行：%s", line)
			}
			continue
		}

		// 解析原子行
		fields := strings.Fields(line)
		if len(fields) != 4 {
			return nil, fmt.Errorf("无效的原子行：%s", line)
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

// DoubleCheck 用于 CalcNMR 检查构象是否合理，以及是否存在重复结构，这是整个 CalcNMR 最核心的步骤
// DoubleCheck 的思路为：
//
//	假设有一个 xyz 文件，其中存在 1000 多个不同构象的原子坐标和能量，
//	循环各个构型，根据其能量和结构相似性判断其是否应当归入某个已有的簇，跟已有的簇没有足够相似性就作为新的簇。
//	能量相似性可以根据能量之间的差值判断，将每一个结构中的能量提取出来组成一个 []float64
//	对于输入文件里第一个结构，或者当前结构和之前已有的所有簇都不相似，那么这个结构将被作为一个新的簇，
//	此簇的能量、结构也等同于这个结构。若当前结构与之前某个簇相似（能量和结构差异都同时小于自设的阈值），
//	那么这个结构就被认为归入了这个簇，因此这个簇的容量会 +1；如果与此同时这个结构的能量比这个簇的能量更低，
//	那么这个结构将被作为这个簇的代表，即使用这个结构的能量和结构作为这个簇的能量和结构。等所有结构都循环完之后，
//	程序就会输出各个簇的绝对能量，相对于能量最低的簇的能量差 (deltaE)，以及这个簇与与它结构最相近的另一个簇的原子间距偏差最大值 (deltaX)。
//	显然，最终得到的簇的数目 <=结构的数目。
//	结构相似性根据原子坐标计算距离矩阵，将其非对角元转化为一维数组形式，并从小到大进行排序，这称为原子间距数组。
//	两个结构的差异用二者的距离间距数组的差值数组的绝对值的最大值来衡量。
//
// @param: thresholds(float): 查找的阈值
// @param: clusters: []Cluster，通过 ParseXyzFile() 方法得到的 []Cluster
func DoubleCheck(thresholds float64, clusters []Cluster) {
	// 用于存储最终的聚类结果
	var finalClusters []Cluster

	// 遍历每一个构型
	for _, currentCluster := range clusters {
		// 标记当前构型是否与已有簇相似
		var foundSimilarCluster bool

		// 遍历已有的簇
		for i, existingCluster := range finalClusters {
			// 判断能量相似性
			energyDiff := math.Abs(currentCluster.Energy - existingCluster.Energy)
			if energyDiff <= thresholds {
				// 判断结构相似性
				if IsClusterSimilar(currentCluster, existingCluster, thresholds) {
					// 将当前构型归入已有簇
					finalClusters[i].Atoms = append(finalClusters[i].Atoms, currentCluster.Atoms...)
					foundSimilarCluster = true
					break
				}
			}
		}

		if !foundSimilarCluster {
			// 将当前构型作为新的簇添加到最终结果中
			finalClusters = append(finalClusters, currentCluster)
		}
	}

	// 输出最终的聚类结果
	for i, cluster := range finalClusters {
		fmt.Printf("Cluster %d:\n", i+1)
		fmt.Printf("Energy: %.2f\n", cluster.Energy)
		fmt.Printf("Atoms:\n")
		for _, atom := range cluster.Atoms {
			fmt.Printf("%s (%.2f, %.2f, %.2f)\n", atom.Symbol, atom.X, atom.Y, atom.Z)
		}
		fmt.Println()
	}

}

// IsClusterSimilar 判断两个簇的结构相似性
func IsClusterSimilar(cluster1, cluster2 Cluster, thresholds float64) bool {
	// 将簇的原子坐标转换为距离数组
	distances1 := CalculateDistances(cluster1.Atoms)
	distances2 := CalculateDistances(cluster2.Atoms)

	// 计算两个距离数组的差值数组
	diff := make([]float64, len(distances1))
	for i := 0; i < len(distances1); i++ {
		diff[i] = math.Abs(distances1[i] - distances2[i])
	}

	// 计算差值数组的最大值
	maxDiff := 0.0
	for _, value := range diff {
		if value > maxDiff {
			maxDiff = value
		}
	}

	// 判断结构相似性
	if maxDiff <= thresholds {
		return true
	}
	return false
}

// CalculateDistances 计算簇中原子的距离数组
func CalculateDistances(atoms []Atom) []float64 {
	distances := make([]float64, 0)
	for i := 0; i < len(atoms); i++ {
		for j := i + 1; j < len(atoms); j++ {
			distance := CalculateDistance(atoms[i], atoms[j])
			distances = append(distances, distance)
		}
	}
	sort.Float64s(distances)
	return distances
}

// CalculateDistance 计算两个原子之间的距离
func CalculateDistance(atom1, atom2 Atom) float64 {
	dx := atom1.X - atom2.X
	dy := atom1.Y - atom2.Y
	dz := atom1.Z - atom2.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}
