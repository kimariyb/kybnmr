package calc

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

/*
* calculate.go
* 该模块主要涉及实现 KYBNMR 运行时所需要的计算功能
*
* @Author: Kimariyb
* @Address: XiaMen University
* @Data: 2023-09-21
 */

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

// ToXYZString 将 cluster 对象转化为 XYZ 坐标
func (c Cluster) ToXYZString() string {
	var sb strings.Builder

	// 写入原子坐标
	for _, atom := range c.Atoms {
		sb.WriteString(fmt.Sprintf("%2s \t\t%14.10f \t\t%14.10f \t\t%14.10f\n", atom.Symbol, atom.X, atom.Y, atom.Z))
	}

	return sb.String()
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
	// 打印一共有多少个 Cluster 同时输出最小的 Cluster 的能量
	fmt.Printf("The total number of clusters searched is: * %d *\n", len(cl))
	fmt.Printf("The lowest energy is: \t%.6f a.u.\n", cl[0].Energy)
	fmt.Println("Sorting clusters according to energy...")
	for i, cluster := range cl {
		// 计算相对能量（以 kcal/mol 为单位）
		relativeEnergy := (cluster.Energy - minEnergy) * 627.51
		// 打印 Cluster 的信息
		fmt.Printf(" # Cluster: %d\tE = %.6f a.u.\tDeltaEnergy = %.2f kcal/mol\n",
			i+1, cluster.Energy, relativeEnergy)
	}
	fmt.Println()
}

// DoubleCheck 用于 KYBNMR 检查构象是否合理，以及是否存在重复结构，这是整个 KYBNMR 最核心的步骤
// 将 clusters 中的第一个 cluster 或者当前 cluster 和 resultClusters 中的所有 cluster 都不相似
// 那么这个 cluster 将被作为一个新的簇，此簇的能量、结构也等同于这个 cluster
// 若当前 cluster 与存在 resultClusters 中的某一个 cluster 相似（能量和结构差异都同时小于自设的阈值），
// 那么这个 cluster 就被认为归入了这个簇，因此这个簇的容量会 +1；
// 如果与此同时这个 cluster 的能量比这个簇的能量更低，那么这个 cluster 将被作为这个簇的代表，
// 即使用这个 cluster 的能量和结构作为这个簇的能量和结构。
// @param: eneThreshold(float): 查找的能量阈值
// @param: disThreshold(float): 查找的距离阈值
// @param: clusters: ClusterList，通过 ParseXyzFile() 方法得到的 ClusterList
// @return: 返回一个 ClusterList
func DoubleCheck(eneThreshold float64, disThreshold float64, clusters ClusterList) (ClusterList, error) {
	// 检查参数有效性
	if eneThreshold < 0 || disThreshold < 0 {
		return nil, errors.New("threshold values must be non-negative")
	}

	if len(clusters) == 0 {
		return nil, errors.New("empty cluster list")
	}

	// 打印 DoubleCheck 运行标志
	fmt.Println()
	fmt.Println("  =======================================")
	fmt.Println("  |             Double Check            |")
	fmt.Println("  =======================================")
	fmt.Println()
	// 创建一个新的切片来存储结果簇
	resultClusters := make(ClusterList, 0)

	// 首先，将第一个 cluster 首先加入 resultClusters 中，作为第一个簇
	resultClusters = append(resultClusters, clusters[0])

	// 接着遍历 clusters 中除第一个以外的每个簇
	for _, cluster := range clusters[1:] {
		// 标识符，默认假设当前簇与已有簇不相似
		isSimilar := false

		// 循环遍历 resultClusters 中的每一个簇
		for i, resultCluster := range resultClusters {
			// 检查当前 clusters 中的簇与 resultClusters 中的每一个簇是否相似
			if IsSimilarToCluster(&cluster, &resultCluster, eneThreshold, disThreshold) {
				// 如果相似，则判断两个 cluster 的能量哪个更小
				isSimilar = true
				// 选择能量更小的簇
				if cluster.Energy < resultCluster.Energy {
					resultClusters[i] = cluster
				}
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

	return resultClusters, nil
}

// IsSimilarToCluster 函数用于检查两个结构是否相似
// 需要同时检查能量差异和结构差异，结构差异使用两原子距离数组的差值数组的绝对值的最大值来衡量。
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

	// 计算距离矩阵
	distMatrix1 := calculateDistanceMatrix(cluster1)
	distMatrix2 := calculateDistanceMatrix(cluster2)

	// 转换为一维数组形式的原子间距数组
	distArray1 := convertToDistanceArray(distMatrix1)
	distArray2 := convertToDistanceArray(distMatrix2)

	// 对距离数组进行排序
	sort.Float64s(distArray1)
	sort.Float64s(distArray2)

	// 计算差值数组的绝对值的最大值
	maxDiff := 0.0
	for i := 0; i < len(distArray1); i++ {
		diff := math.Abs(distArray1[i] - distArray2[i])
		if diff > maxDiff {
			maxDiff = diff
		}
	}

	// 如果结构差异超过阈值，则认为两个结构不相似
	if maxDiff > disThreshold {
		return false
	}

	// 两个结构相似
	return true
}

// calculateDistanceMatrix 函数用于计算距离矩阵
func calculateDistanceMatrix(cluster *Cluster) [][]float64 {
	numAtoms := len(cluster.Atoms)
	distMatrix := make([][]float64, numAtoms)

	// 遍历每对原子，计算它们之间的距离并存储在距离矩阵中
	for i := 0; i < numAtoms; i++ {
		distMatrix[i] = make([]float64, numAtoms)
		for j := 0; j < numAtoms; j++ {
			if i == j {
				distMatrix[i][j] = 0.0 // 对角元素为 0，表示同一个原子
			} else {
				// 计算两个原子在X、Y、Z轴上的坐标差异
				diffX := cluster.Atoms[i].X - cluster.Atoms[j].X
				diffY := cluster.Atoms[i].Y - cluster.Atoms[j].Y
				diffZ := cluster.Atoms[i].Z - cluster.Atoms[j].Z

				// 计算两个原子之间的距离并存储在距离矩阵中
				distMatrix[i][j] = math.Sqrt(diffX*diffX + diffY*diffY + diffZ*diffZ)
			}
		}
	}

	return distMatrix
}

// convertToDistanceArray 函数将距离矩阵转换为一维数组形式的原子间距数组
func convertToDistanceArray(distMatrix [][]float64) []float64 {
	numAtoms := len(distMatrix)
	distArray := make([]float64, numAtoms*(numAtoms-1)/2)

	index := 0
	// 遍历距离矩阵的非对角元素，将它们存储在一维数组中
	for i := 0; i < numAtoms; i++ {
		for j := i + 1; j < numAtoms; j++ {
			distArray[index] = distMatrix[i][j]
			index++
		}
	}

	return distArray
}
