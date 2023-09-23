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
	PreOptArgs    string
	PostOptArgs   string
	PreThreshold  string
	PostThreshold string
	Gaupath       string
	Orcapath      string
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
	optConfig.PreThreshold = optimizedSection.Key("preThreshold").String()
	optConfig.PostThreshold = optimizedSection.Key("postThreshold").String()
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

// ClusterList 定义 ClusterList 类型
type ClusterList []Cluster

// SortCluster 方法对 ClusterList 中的所有 Cluster 进行排序
// 遍历 ClusterList 中的每一个 Cluster，按照能量从低到高排序
func (cl ClusterList) SortCluster() {
	sort.SliceStable(cl, func(i, j int) bool {
		return cl[i].Energy < cl[j].Energy
	})
}

// PrintClusterInFo 方法可以按顺序打印 ClusterList 中所有 Cluster 的信息
// 同时输出绝对能量 (Hartree)，以及与能量最小的 Cluster 的相对能量 (kcal/mol)，以及与结构阈值的差值
// 打印的格式如下：
// #  Cluster: 1  E = -44.774700 a.u.  DeltaDistance = 0.11  DeltaEnergy = 0.00 kcal/mol
// #  Cluster: 2  E = -44.774697 a.u.  DeltaDistance = 0.11  DeltaEnergy = 0.00 kcal/mol
func (cl ClusterList) PrintClusterInFo() {
	// 首先对 ClusterList 中的 Cluster 进行排序
	cl.SortCluster()
	// 获取能量最小的 Cluster
	minEnergy := cl[0].Energy
	for i, cluster := range cl {
		// 计算相对能量（以 kcal/mol 为单位）
		relativeEnergy := (cluster.Energy - minEnergy) * 627.5094

		// 计算与结构阈值的差值
		disThreshold := 0.11 // 结构阈值（根据实际情况进行调整）
		deltaDistance := math.Abs(cluster.Energy - disThreshold)

		// 打印 Cluster 的信息
		fmt.Printf("#  Cluster: %d  E = %.6f a.u.  DeltaDistance = %.2f  DeltaEnergy = %.2f kcal/mol\n",
			i+1, cluster.Energy, deltaDistance, relativeEnergy)
	}
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
			energy = 0.0
			continue
		}

		if lineCount == 2 {
			// 解析能量行
			energy, err = strconv.ParseFloat(strings.TrimSpace(line), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid energy rows：%s", line)
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

// WriteXyzFile 向新的 xyz 文件中写入信息
// @param clusters: []Cluster 需要写入的文件信息
// @param xyzFileName: string 需要写入的 xyz 文件的名称
func WriteXyzFile(clusters []Cluster, xyzFileName string) {
	//
}

// DoubleCheck 用于 CalcNMR 检查构象是否合理，以及是否存在重复结构，这是整个 CalcNMR 最核心的步骤
//
// 能量和几何结构偏差作为判断两个结构是否相似的标准，同时小于两个阈值才算相似
// 能量差异偏差的计算方式为比较两者的能量差值，因为比较的单位为 kcal/mol，因此在两者做差之后，还需要乘上 627.5094 才能和阈值对比
// 如果两者的差值比阈值小，则满足能量标准；如果两者的差值比阈值大，则不满足能量标准。
// 结构差异偏差的计算方式为矩阵计算方法，首先根据原子坐标计算距离矩阵，将其非对角元转化为一维数组形式，并从小到大进行排序，这称为原子间距数组。
// 两个结构之间的结构差异用二者的距离间距数组的差值数组的绝对值的最大值来衡量。
// 如果两者距离间距数组的差值数组的绝对值的最大值比阈值小，则满足结构标准；如果比阈值大，则不满足结构标准。
//
// 1. 首先循环遍历每一个结构，对于输入文件里第一个结构，或者当前结构和之前已有的所有簇都不相似，那么这个结构将被作为一个新的簇，此簇的能量、结构也等同于这个结构。
// 2. 若当前结构与之前某个簇相似（能量和结构差异都同时小于自设的阈值），那么这个结构就被认为归入了这个簇，因此这个簇的容量会 +1。
// 3. 如果与此同时这个结构的能量比这个簇的能量更低，那么这个结构将被作为这个簇的代表，即使用这个结构的能量和结构作为这个簇的能量和结构。
// 4. 等所有结构都循环完之后，程序就会输出各个簇的绝对能量，相对于能量最低的簇的能量差(DeltaE)，以及这个簇与与它结构最相近的另一个簇的原子间距偏差最大值(DeltaDis)。
// 5. 最后把所有簇都存储在一个 []Cluster 中，同时返回这个 []Cluster
//
// @param: eneThreshold(float): 查找的能量阈值
// @param: disThreshold(float): 查找的距离阈值
// @param: clusters: ClusterList，通过 ParseXyzFile() 方法得到的 ClusterList
// @return: 返回一个 ClusterList
func DoubleCheck(eneThreshold float64, disThreshold float64, clusters ClusterList) ClusterList {
	// 创建一个新的切片来存储结果簇
	resultClusters := make(ClusterList, 0)

	// 遍历clusters中的每个簇
	for _, cluster := range clusters {
		// 默认假设当前簇与已有簇不相似
		isSimilar := false

		// 检查当前簇与已有簇是否相似
		for _, resultCluster := range resultClusters {
			if IsSimilarToCluster(&cluster, &resultCluster, eneThreshold, disThreshold) {
				isSimilar = true
				break
			}
		}

		// 如果当前簇与已有簇不相似，则将其添加到结果簇中
		if !isSimilar {
			resultClusters = append(resultClusters, cluster)
		}
	}

	// 打印 resultClusters 的信息
	resultClusters.PrintClusterInFo()

	return resultClusters
}

// IsSimilarToCluster 函数用于检查两个结构是否相似
// 参数 cluster1 和 cluster2 分别是待比较的两个 Cluster 指针
// 参数 eneThreshold 是能量差异的阈值（以 Hartree 为单位）
// 参数 disThreshold 是结构差异的阈值（以 Angstrom 为单位）
// 函数返回一个布尔值，表示两个结构是否相似
func IsSimilarToCluster(cluster1, cluster2 *Cluster, eneThreshold, disThreshold float64) bool {
	// 检查能量差异
	// 计算 cluster1 和 cluster2 的能量差异，并将其转换为以 kcal/mol 为单位
	eneDiff := math.Abs(cluster1.Energy-cluster2.Energy) * 627.5094

	// 如果能量差异超过阈值，则认为两个结构不相似
	if eneDiff > eneThreshold {
		return false
	}

	// 检查结构差异
	// 计算 cluster1 和 cluster2 的原子间距离数组
	disArray1 := calculateDistanceArray(cluster1.Atoms)
	disArray2 := calculateDistanceArray(cluster2.Atoms)

	// 记录最大的距离差异
	maxDiff := 0.0

	// 遍历原子间距离数组，计算距离差异的最大值
	for i := range disArray1 {
		diff := math.Abs(disArray1[i] - disArray2[i])
		if diff > maxDiff {
			maxDiff = diff
		}
	}

	// 如果距离差异超过阈值，则认为两个结构不相似
	if maxDiff > disThreshold {
		return false
	}

	// 两个结构相似
	return true
}

// calculateDistanceArray 函数计算原子间距离数组
// 参数 geometry 是一个包含多个 Atom 的切片，表示一组原子的几何结构
// 函数返回一个包含所有原子间距离的一维数组
func calculateDistanceArray(geometry []Atom) []float64 {
	// 计算原子数量
	n := len(geometry)

	// 计算原子间距离数组的长度
	// 假设有 m 个原子，那么原子间距离数组的长度为 m*(m-1)/2
	disArray := make([]float64, n*(n-1)/2)

	// disArray 数组索引
	idx := 0

	// 遍历所有原子对，计算它们之间的距离并存储到 disArray 中
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			disArray[idx] = calculateDistance(geometry[i], geometry[j])
			idx++
		}
	}

	// 返回原子间距离数组
	return disArray
}

// calculateDistance 计算两个原子之间的距离
func calculateDistance(atom1, atom2 Atom) float64 {
	dx := atom1.X - atom2.X
	dy := atom1.Y - atom2.Y
	dz := atom1.Z - atom2.Z

	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}
