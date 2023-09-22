package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

/*
* common.go
* 该模块是 CalcNMR 程序所需要使用到的工具函数等
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
     |                          CalcNMR                          |
     |                   =====================                   |
     |                        Kimari Y.B.                        |
     |        School of Electronic Science and Engineering       |
     |                     XiaMen University                     |
      -----------------------------------------------------------
      * CalcNMR version ` + version + ` on ` + date + `
      * Homepage is https://github.com/kimariyb/CalcNMR
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
// 首先扫描程序运行的文件夹目录中的所有文件
// 将除开 CalcNMR 主程序、xyz 文件、xtb.trj 文件、ini 文件 都移动到 temp 文件夹中
func RemoveTempFolder() {
	// 获取当前目录文件夹
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}
	// 需要保留的文件类型和名称列表
	keepFiles := []string{"CalcNMR", "*.xyz", "xtb.trj", "*.ini"}

	// 遍历文件夹中的所有文件
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 判断文件是否为目录
		if info.IsDir() {
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
