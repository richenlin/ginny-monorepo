package util

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// Exists 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// IsDir 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

// MkDir 创建目录
func MkDir(path string) error {
	// 创建文件夹
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// ReadFile 读取文件
func ReadFile(path string) ([]byte, error) {
	if !Exists(path) {
		return nil, errors.New("file not exists")
	}

	bin, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("read file failed")
	}

	return bin, nil
}

// RemoveFile 删除文件
func RemoveFile(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

// CopyFile 拷贝文件
func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dstPath := path.Dir(dst)
	if !Exists(dstPath) {
		if err := MkDir(dstPath); err != nil {
			return err
		}
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

// ReplaceFile 替换文件内容
func ReplaceFile(filePath string, origin, target string) error {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		return err
	}
	defer f.Close()
	br := bufio.NewReader(f)
	output := make([]byte, 0)

	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if ok, _ := regexp.Match(origin, line); ok {
			reg := regexp.MustCompile(origin)
			newByte := reg.ReplaceAll(line, []byte(target))
			output = append(output, newByte...)
			output = append(output, []byte("\n")...)
		} else {
			output = append(output, line...)
			output = append(output, []byte("\n")...)
		}
	}

	return WriteToFile(filePath, output)
}

// FileHasContainsStr
func FileHasContainsStr(filepath, target string) (bool, error) {
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		return false, err
	}
	defer f.Close()
	br := bufio.NewReader(f)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, err
		}
		if ok, _ := regexp.Match(target, line); ok {
			return true, nil
		}
	}
	return false, nil
}

// WriteToFile 回写文件内容
func WriteToFile(filePath string, outPut []byte) error {
	var (
		f   *os.File
		err error
	)
	if !Exists(filePath) {
		dir := filepath.Dir(filePath)
		if !Exists(dir) {
			_ = MkDir(dir)
		}
		f, err = os.Create(filePath)
	} else {
		f, err = os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0600)
	}

	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(f)
	_, err = writer.Write(outPut)
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}

// GetFiles 获取目录所有文件
func GetFiles(path string) []string {
	files := make([]string, 0)
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			if strings.HasPrefix(f.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
	return files
}

// fileName:文件名字(带全路径)
// content: 写入的内容
func AppendToFile(fileName string, content []byte) error {
	// 以只写的模式，打开文件
	f, err := os.OpenFile(fileName, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("cacheFileList.yml file create failed. err: " + err.Error())
	} else {
		// 查找文件末尾的偏移量
		n, _ := f.Seek(0, os.SEEK_END)
		// 从末尾的偏移量开始写入内容
		_, err = f.WriteAt(content, n)
	}
	defer func() {
		_ = f.Close()
	}()
	return err
}

// GetStrIndex 获取指定字符串所在行
func GetStrIndex(path, patten string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	// ma := regexp.MustCompile(`^plugins:$`)
	ma := regexp.MustCompile(patten)
	// Splits on newlines by default.
	scanner := bufio.NewScanner(f)
	line := 1
	// https://golang.org/pkg/bufio/#Scanner.Scan
	for scanner.Scan() {
		if ma.MatchString(scanner.Text()) {
			return line, nil
		}
		line++
	}
	return 0, nil
}

// ReadLine 读取指定范围行内容
func ReadLine(path string, start, end int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	result := ""
	fileScanner := bufio.NewScanner(f)
	lineCount := 1
	for fileScanner.Scan() {
		if lineCount >= start {
			if end > 0 && end > lineCount {
				break
			}
			result += fileScanner.Text() + "\n"
		}
		lineCount++
	}

	return result, nil
}
