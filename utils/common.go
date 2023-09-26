package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

/*
* common.go
* 该模块是 KYBNMR 程序所需要使用到的工具函数等
*
* @Method:
*	ShowHead(): 展示程序头，包括作者信息、版本信息等
*
* @Author: Kimariyb
* @Address: XiaMen University
* @Data: 2023-09-21
 */

// ShowHead 展示程序头，包括作者信息、版本信息等
// @Return: version(string): 版本号
// @Return: date(string): 发行时间
func ShowHead() (string, string) {
	version := "1.0.0(dev)"
	date := "2023-09-21"

	asciiArt := `
      -----------------------------------------------------------
     |                   =====================                   |
     |                          KYBNMR                          |
     |                   =====================                   |
     |                        Kimari Y.B.                        |
     |        School of Electronic Science and Engineering       |
     |                     XiaMen University                     |
      -----------------------------------------------------------
      * KYBNMR version ` + version + ` on ` + date + `
      * Homepage is https://github.com/kimariyb/kybnmr
	`

	fmt.Println(asciiArt)

	return version, date
}

// CheckFileCurrentExist 检查当前运行脚本的目录下是否存在指定文件
// 如果存在，则返回 true，同时返回 filename 的绝对路径
// 如果不存在，则返回 false
func CheckFileCurrentExist(filename string) (bool, string) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		fmt.Println("Error getting absolute path:", err)
		return false, ""
	}

	_, err = os.Stat(absPath)
	if os.IsNotExist(err) {
		return false, ""
	}

	return true, absPath
}

// CheckFileType 检查文件的类型是否为指定类型
// @Param: filename(string): 文件名
// @Param: fileType(string): 指定的文件类型（后缀名），例如 ".txt"
// @Return: bool
func CheckFileType(filename, fileType string) bool {
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return false
	}

	if fileInfo.IsDir() {
		fmt.Println("Error: Input is a directory, not a file")
		return false
	}

	// 拿到文件的扩展名
	extension := strings.ToLower(fileInfo.Name()[strings.LastIndex(fileInfo.Name(), "."):])
	if extension != fileType {
		fmt.Println("Error: File type is not", fileType)
		return false
	}

	return true
}

// RemoveTempFolder 将文件移动到临时文件夹 temp 中
// 扫描程序运行的文件夹目录中的所有文件，将指定文件都移动到 temp 文件夹中
// 不移动目录下的任何文件夹，以及文件夹中的文件
func RemoveTempFolder(keepFiles []string) {
	// 获取当前目录文件夹
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	// 遍历文件夹中的所有文件
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 判断文件是否为目录
		if info.IsDir() {
			// 判断是否当前目录
			if path != dir {
				// 忽略子目录，直接跳过
				return filepath.SkipDir
			}
			return nil
		}

		// 判断文件是否需要保留
		keep := false
		for _, pattern := range keepFiles {
			match, err := filepath.Match(pattern, info.Name())
			if err != nil {
				return err
			}
			if match {
				keep = true
				break
			}
		}

		// 如果文件不需要保留，则移动到 temp 文件夹中
		if !keep {
			newPath := filepath.Join("temp", info.Name())
			err := os.Rename(path, newPath)
			if err != nil {
				fmt.Println("Error moving file:", err)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error walking directory:", err)
		return
	}
}

// RemoveOutFile 移动 out 文件至指定文件夹
func RemoveOutFile(targetFolder string, ) {

}

// RenameFile 工具函数，修改文件的名字
// @param: olderFileName(string): 旧的文件名
// @param: newFileName(string): 新的文件名
func RenameFile(olderFileName string, newFileName string) {
	err := os.Rename(olderFileName, newFileName)
	if err != nil {
		fmt.Println("Error renaming file:", err)
		return
	}
	fmt.Println("File renamed successfully.")
}

// SplitStringBySpace 根据一段字符串的空格，切割字符串，并存在一个 string[] 中，同时返回
func SplitStringBySpace(str string) []string {
	return strings.Split(str, " ")
}

// SplitStringByComma 根据一段字符串的 "," 切割字符串，并返回成一个 float64[]
func SplitStringByComma(str string) []float64 {
	valuesStr := strings.Split(str, ",")
	values := make([]float64, len(valuesStr))
	for i, v := range valuesStr {
		val, _ := strconv.ParseFloat(strings.TrimSpace(v), 64)
		values[i] = val
	}
	return values
}

// FormatDuration 将时间转化为特定的的格式，同时输出当前时间
func FormatDuration(duration time.Duration) {
	fmt.Println()
	fmt.Println("----------------------------------------------------------------")
	fmt.Println("Thanks for your use !!!")
	durationString := fmt.Sprintf("%02dh : %02dm : %02ds", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60) // 输出 KYBNMR 程序运行的总时间
	fmt.Printf("Time spent running KYBNMR: %s\n", durationString)
	// 展示结束语，同时输出当前时间日期，以及版权 (c)
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("KYBNMR finished at %s. Copyright (c) Kimariyb\n", currentTime)
	fmt.Println("----------------------------------------------------------------")
	fmt.Println()
}

// ParseSkipSteps 解析 skipSteps 字符串为整数数组
func ParseSkipSteps(skipSteps string) []int {
	var steps []int

	if skipSteps != "" {
		stepStrs := strings.Split(skipSteps, " ")
		for _, stepStr := range stepStrs {
			step, err := strconv.Atoi(stepStr)
			if err != nil {
				fmt.Println("Invalid skip step:", stepStr)
			} else {
				steps = append(steps, step)
			}
		}
	}

	return steps
}

// ContainsArray 判断整数数组中是否包含指定的值
func ContainsArray(array []int, value int) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}
