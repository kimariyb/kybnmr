package calc

import (
	"bufio"
	"fmt"
	"gopkg.in/ini.v1"
	"kybnmr/utils"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

/*
* config.go
* 该模块主要涉及实现 KYBNMR 运行时所需要读取配置、写入文件功能
*
*	Config 结构体，用来存储 ini 文件内的配置。在 KYBNMR 的 ini 文件中，你可以配置以下属性
* 	[dynamics] 使用 xtb 做动力学的配置项
*		temperature(float): 温度，单位为 K
*		time(float): 时间，单位为 ps
*		dump(float): 每间隔多少时间往轨迹文件里写入一次，单位为 fs
*		step(float): 步长，单位为 fs
*		velo(bool): 是否输出速度
*		nvt(bool): 是否开启 NVT 模拟
*		hmass(int): 氢原子的质量是实际的多少倍
*		shake(int): 将与氢有关的化学键距离都用 SHAKE 算法约束住
*		sccacc(float): 动态 xTB 计算的准确性
*		dynamicArgs(string): 运行 xtb 做动力学的命令
*
*	[optimized] 使用 xtb 做预优化的配置项、使用 Gaussian 和 orca 做进一步优化的配置项
*		preOptArgs(string)：预优化的参数
*		postOptArgs(string): 进一步优化的参数
*		preThreshold(string): 预优化之后的阈值
*		postThreshold(string): 进一步优化之后的阈值
*		gauPath(string): gaussian 运行路径
*		orcaPath(string): orca 运行路径
*		shermoPath(string): shermo 运行路径
*
* @Author: Kimariyb
* @Address: XiaMen University
* @Data: 2023-09-21
 */

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
	ShermoPath    string
}

// Config 记录 ini 文件配置类
type Config struct {
	DyConfig  DynamicsConfig
	OptConfig OptimizedConfig
}

type ShermoResult struct {
	FileName string
	Energy   string
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
	// 分别解析 ini 文件中的 [dynamics]、[optimized]组分别存储在
	// DynamicsConfig、OptimizedConfig 结构体中
	// 最后将 DynamicsConfig、OptimizedConfig 结构体存储在 Config 中
	dynamicsSection := iniFile.Section("dynamics")
	optimizedSection := iniFile.Section("optimized")

	// 声明一个 dynamicsConfig、OptimizedConfig
	dynamicsConfig := DynamicsConfig{}
	optConfig := OptimizedConfig{}

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
	optConfig.ShermoPath = optimizedSection.Key("shermoPath").String()

	// 给 config 赋值
	config.DyConfig = dynamicsConfig
	config.OptConfig = optConfig

	return config
}

// ParseOutFile 解析 out 文件，将最后一帧的结构保存在 Cluster 中
//   - softwareName: 使用的是 orca 还是 gaussian 程序生成的 out 文件
//   - filePath: 需要解析的 out 文件的路径
func ParseOutFile(softwareName string, filePath string) (Cluster, error) {
	var cluster Cluster
	var err error

	// 首先判断 filePath 是否为一个 out 文件
	if !utils.CheckFileType(filePath, ".out") {
		// 如果不是 out 文件则直接退出并报错
		return Cluster{}, fmt.Errorf("error the format of input file")
	}

	// 判断 softwareName 传入的参数是 orca 还是 gaussian
	if strings.EqualFold(softwareName, "orca") {
		// 如果传入的是 orca 则调用 parseOrcaOutput()
		cluster, err = parseOrcaOutput(filePath)
		if err != nil {
			return Cluster{}, err
		}
	} else if strings.EqualFold(softwareName, "gaussian") {
		// 如果传入的是 gaussian 则调用 parseGauOutput()
		cluster, err = parseGauOutput(filePath)
		if err != nil {
			return Cluster{}, err
		}
	}

	// 返回一个 Cluster 对象
	return cluster, nil
}

// ParseGauOutput 读取 Gaussian 生成的 out 文件
// 在 Gaussian 生成的 out 文件的最后，都会出现一个 Standard orientation
// 在 Standard orientation 表格中以下变量至关重要
// Atomic Number: 原子序号，代表元素周期表中的位置，1 代表 H；2 代表 He 以此类推
// Coordinates (Angstroms): 原子坐标（单位为埃），分为 X；Y；Z 坐标，这是一个笛卡尔坐标系的坐标
//
//	Standard orientation:
//
// ---------------------------------------------------------------------
// Center     Atomic      Atomic             Coordinates (Angstroms)
// Number     Number       Type             X           Y           Z
// ---------------------------------------------------------------------
//
//	 1          8           0        1.169391   -0.453770   -0.882827
//	 2          8           0       -1.184882    2.809963   -0.433427
//	 3          8           0        0.716726    2.116662    0.491823
//	 4          8           0        3.366765   -0.337316   -0.613653
//	 5          6           0       -0.108179   -0.650924   -0.403350
//	 6          6           0       -0.916481    0.439636   -0.034821
//	 7          6           0       -0.637019   -1.942534   -0.425943
//	 8          6           0       -2.250913    0.196603    0.327706
//	 9          6           0       -1.961017   -2.165935   -0.050206
//	10          6           0       -2.771332   -1.095676    0.332879
//	11          6           0       -0.362052    1.832310    0.029385
//	12          6           0        2.305426   -0.474834   -0.073872
//	13          6           0        2.087999   -0.671809    1.405520
//	14          1           0        0.000372   -2.757450   -0.754757
//	15          1           0       -2.875364    1.028823    0.643617
//	16          1           0       -2.359654   -3.176253   -0.063447
//	17          1           0       -3.801283   -1.265085    0.631546
//	18          1           0        1.568720    0.203757    1.806039
//	19          1           0        3.063572   -0.771149    1.881978
//	20          1           0        1.481084   -1.558556    1.612610
//	21          1           0       -1.940041    2.410593   -0.896697
//
// ---------------------------------------------------------------------
func parseGauOutput(filePath string) (Cluster, error) {
	var nAtoms int
	var foundLastOrientation bool
	var atoms []Atom
	var lastOrientationAtoms []Atom

	// 首先将扫描到的文件变为绝对路径，再打开文件
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return Cluster{}, err
	}

	file, err := os.Open(absPath)
	if err != nil {
		return Cluster{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// 首先扫描 out 文件，在 out 文件中找到 NAtoms= 随便读取一个后面的数字，例
	// 如读取 NAtoms=  21 中的 21
	// 将这个数字赋值给 nAtoms 变量
	nAtoms, err = extractNAtomsFromFile(absPath)
	if err != nil {
		return Cluster{}, err
	}

	for scanner.Scan() {
		line := scanner.Text()
		// 接着找到文件中最后一个 Standard orientation
		// 定位到最后一个 Standard orientation 一行后，接着跳过四行。因为后面四行为表格线
		if strings.Contains(line, "Standard orientation") {
			// 在找到新的 "Standard orientation" 时，将上一次找到的最后一个 "Standard orientation" 对应的 atoms 清空
			atoms = atoms[:0]
			lastOrientationAtoms = make([]Atom, 0) // 初始化为新的空切片
			foundLastOrientation = false
			// 跳过四行表格线
			for i := 0; i < 4; i++ {
				scanner.Scan()
			}
		}

		if strings.Contains(line, "Standard orientation") {
			foundLastOrientation = true
		}

		// 到第一行时 1  8  0  1.169391   -0.453770   -0.882827
		// 只需要关注第二个列的 8 和后三列的 x、y、z 坐标。其中 8 代表是第八个元素氧。
		// 将第一行的原子坐标和元素保存为一个 Atom 结构体
		if foundLastOrientation && len(atoms) < nAtoms {
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				atomicNumber, err := strconv.Atoi(fields[1])
				if err != nil {
					return Cluster{}, fmt.Errorf("unable to resolve atomic number: %s", fields[1])
				}
				symbol, err := getSymbol(atomicNumber)
				if err != nil {
					return Cluster{}, err
				}
				x, err := strconv.ParseFloat(fields[3], 64)
				if err != nil {
					return Cluster{}, fmt.Errorf("unable to resolve X-coordinate: %s", fields[3])
				}
				y, err := strconv.ParseFloat(fields[4], 64)
				if err != nil {
					return Cluster{}, fmt.Errorf("unable to resolve Y-coordinate: %s", fields[4])
				}
				z, err := strconv.ParseFloat(fields[5], 64)
				if err != nil {
					return Cluster{}, fmt.Errorf("unable to resolve Z-coordinate: %s", fields[5])
				}

				atoms = append(atoms, Atom{
					Symbol: symbol,
					X:      x,
					Y:      y,
					Z:      z,
				})

				if foundLastOrientation && len(atoms) > 0 {
					lastOrientationAtoms = append(lastOrientationAtoms, atoms...)
				}
			}
		}
	}

	cluster := Cluster{
		Atoms:  atoms,
		Energy: 0,
	}

	if err := scanner.Err(); err != nil {
		return Cluster{}, fmt.Errorf("error while reading file: %v", err)
	}

	// 接下来扫描 nAtoms 行，每一行的操作都和第一行一样。将所有的 Atom 结构体都赋值给 Cluster 结构体
	// 所有的能量都赋值为 0
	return cluster, nil
}

func extractNAtomsFromFile(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 使用正则表达式匹配 NAtoms= 后面的数字
		re := regexp.MustCompile(`NAtoms=\s*(\d+)`)
		match := re.FindStringSubmatch(line)
		if len(match) > 1 {
			nAtomsStr := match[1]
			nAtoms, err := strconv.Atoi(nAtomsStr)
			if err != nil {
				return 0, fmt.Errorf("unable to convert NAtoms to integer: %s", nAtomsStr)
			}
			return nAtoms, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("failed to scan file: %v", err)
	}

	return 0, fmt.Errorf("NAtoms not found in the file")
}

// getSymbol 根据原子序数获取元素符号
func getSymbol(atomicNumber int) (string, error) {
	// 这里仅对元素周期表的前 100 个元素进行映射
	symbolMap := map[int]string{
		1:   "H",
		2:   "He",
		3:   "Li",
		4:   "Be",
		5:   "B",
		6:   "C",
		7:   "N",
		8:   "O",
		9:   "F",
		10:  "Ne",
		11:  "Na",
		12:  "Mg",
		13:  "Al",
		14:  "Si",
		15:  "P",
		16:  "S",
		17:  "Cl",
		18:  "Ar",
		19:  "K",
		20:  "Ca",
		21:  "Sc",
		22:  "Ti",
		23:  "V",
		24:  "Cr",
		25:  "Mn",
		26:  "Fe",
		27:  "Co",
		28:  "Ni",
		29:  "Cu",
		30:  "Zn",
		31:  "Ga",
		32:  "Ge",
		33:  "As",
		34:  "Se",
		35:  "Br",
		36:  "Kr",
		37:  "Rb",
		38:  "Sr",
		39:  "Y",
		40:  "Zr",
		41:  "Nb",
		42:  "Mo",
		43:  "Tc",
		44:  "Ru",
		45:  "Rh",
		46:  "Pd",
		47:  "Ag",
		48:  "Cd",
		49:  "In",
		50:  "Sn",
		51:  "Sb",
		52:  "Te",
		53:  "I",
		54:  "Xe",
		55:  "Cs",
		56:  "Ba",
		57:  "La",
		58:  "Ce",
		59:  "Pr",
		60:  "Nd",
		61:  "Pm",
		62:  "Sm",
		63:  "Eu",
		64:  "Gd",
		65:  "Tb",
		66:  "Dy",
		67:  "Ho",
		68:  "Er",
		69:  "Tm",
		70:  "Yb",
		71:  "Lu",
		72:  "Hf",
		73:  "Ta",
		74:  "W",
		75:  "Re",
		76:  "Os",
		77:  "Ir",
		78:  "Pt",
		79:  "Au",
		80:  "Hg",
		81:  "Tl",
		82:  "Pb",
		83:  "Bi",
		84:  "Po",
		85:  "At",
		86:  "Rn",
		87:  "Fr",
		88:  "Ra",
		89:  "Ac",
		90:  "Th",
		91:  "Pa",
		92:  "U",
		93:  "Np",
		94:  "Pu",
		95:  "Am",
		96:  "Cm",
		97:  "Bk",
		98:  "Cf",
		99:  "Es",
		100: "Fm",
	}

	symbol, ok := symbolMap[atomicNumber]
	if !ok {
		return "", fmt.Errorf("unknown atomic number: %d", atomicNumber)
	}
	return symbol, nil
}

// parseOrcaOutput 读取 Orca 生成的 out 文件
func parseOrcaOutput(filePath string) (Cluster, error) {
	var cluster Cluster
	return cluster, nil
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
