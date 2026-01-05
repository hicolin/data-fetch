package env

import (
	"os"
	"path/filepath"
)

var RootDir string

func Init() {
	var err error
	RootDir, err = os.Getwd()
	if err != nil {
		panic("无法获取当前工作目录: " + err.Error())
	}

	for {
		modFile := filepath.Join(RootDir, "go.mod")
		if _, err := os.Stat(modFile); err == nil {
			return
		}

		parentDir := filepath.Dir(RootDir)

		if parentDir == RootDir {
			panic("在所有父目录中都未找到 go.mod 文件。请确保在项目目录或其子目录中运行程序。")
		}

		RootDir = parentDir
	}
}
