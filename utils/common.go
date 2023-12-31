package utils

import (
	"fmt"
	"io/fs"
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
     |                          KYBNMR                           |
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

// MoveFile 移动文件
// 将源文件移动到目标路径，并删除源文件
// 参数：
//   - sourcePath：源文件路径
//   - destPath：目标文件路径
//
// 返回值：
//   - error：如果移动文件过程中发生错误，则返回相应的错误信息；否则返回 nil
func MoveFile(sourcePath, destPath string) error {
	err := os.Rename(sourcePath, destPath)
	if err != nil {
		return err
	}
	return nil
}

// MoveFileForType 移动当前文件夹下的所有某一类型的文件至指定文件夹
// 不移动目录下的任何文件夹，以及文件夹中的文件
// 可以选择不移动指定的某一文件或多个文件
func MoveFileForType(fileType string, targetFolder string, excludeFiles ...string) {
	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}

	// 遍历当前文件夹中的文件
	err = filepath.WalkDir(currentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 如果当前路径是文件夹，则跳过
		if d.IsDir() {
			// 判断是否当前目录
			if path != currentDir {
				// 忽略子目录，直接跳过
				return filepath.SkipDir
			}
			return nil
		}

		// 获取文件的扩展名
		fileExt := strings.ToLower(filepath.Ext(path))
		// 如果文件的扩展名与目标类型匹配，并且文件名不在排除列表中，则移动文件
		if fileExt == strings.ToLower(fileType) && !contains(excludeFiles, d.Name()) {
			// 构建目标文件路径
			destPath := filepath.Join(targetFolder, d.Name())
			// 移动文件
			err := MoveFile(path, destPath)
			if err != nil {
				fmt.Printf("Failed to move file: %s\n", err)
			} else {
				fmt.Printf("Move file successfully: %s\n", path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %s\n", err)
	}
}

// 辅助函数：检查字符串切片中是否包含指定的字符串
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// MoveAllFileButKeepFile
// 扫描程序运行的文件夹目录中的所有文件，将除了指定文件之外的文件都移动到指定文件夹中
// 不移动目录下的任何文件夹，以及文件夹中的文件
func MoveAllFileButKeepFile(keepFiles []string, targetFolder string) {
	// 获取当前目录文件夹
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	// 遍历文件夹中的所有文件
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 判断文件是否为目录
		if d.IsDir() {
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
			match, err := filepath.Match(pattern, d.Name())
			if err != nil {
				return err
			}
			if match {
				keep = true
				break
			}
		}

		// 如果文件不需要保留，则移动到指定文件夹中
		if !keep {
			newPath := filepath.Join(targetFolder, d.Name())
			err := MoveFile(path, newPath)
			if err != nil {
				fmt.Printf("Failed to move file: %s\n", err)
			} else {
				fmt.Printf("Move file successfully: %s\n", path)
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

// DeleteAllFileButKeepType
// 删除当前运行文件夹的 thermo/opt 和 thermo/sp 文件夹中的
// 除指定文件类型 keepType 之外的所有文件
// 不删除 thermo/opt 和 thermo/sp 中的文件夹
func DeleteAllFileButKeepType(keepType string) {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current working directory:", err)
		return
	}

	// 定义要处理的文件夹
	folders := []string{
		filepath.Join(currentDir, "thermo", "opt"),
		filepath.Join(currentDir, "thermo", "sp"),
	}

	// 遍历文件夹
	for _, folder := range folders {
		// 打开文件夹
		dir, err := os.Open(folder)
		if err != nil {
			fmt.Println("Failed to open folder:", folder, err)
			continue
		}
		defer dir.Close()

		// 读取文件夹中的文件
		files, err := dir.Readdir(-1)
		if err != nil {
			fmt.Println("Failed to read folder contents:", folder, err)
			continue
		}

		// 遍历文件
		for _, file := range files {
			// 跳过子文件夹
			if file.IsDir() {
				continue
			}

			// 检查文件类型是否匹配指定类型
			if !strings.HasSuffix(file.Name(), keepType) {
				// 删除文件
				filePath := filepath.Join(folder, file.Name())
				err := os.Remove(filePath)
				if err != nil {
					fmt.Println("Failed to delete file:", filePath, err)
				} else {
					fmt.Println("Deleted file:", filePath)
				}
			}
		}
	}
}
