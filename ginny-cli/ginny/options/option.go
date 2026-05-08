package options

import (
	"fmt"
)

var (
	Logo = `

	┌─┐┬┌┐┌┌┐┌┬ ┬
	│ ┬│││││││└┬┘
	└─┘┴┘└┘┘└┘ ┴ 
	
	`
	// 模板仓库地址
	TemplateRepo = "https://github.com/goriller/ginny-template.git"
	// 组件模板仓库地址
	ComponentTemplateRepo = "https://github.com/goriller/ginny-component-template.git"

	// 项目标识
	ProjectFlag = ".ginny.yml"
	// 组件临时缓存目录
	TempPath = ".ginny"
	// 项目proto文件目录
	ProtoPath = "%s/api/proto/main.proto"
	// 替换锚点
	AnchorFlag           = "锚点请勿删除! Do not delete this line!"
	ReplaceAnchorMap     = map[string]interface{}{}
	RegReplaceWithAnchor = func(rpsType string, fn func() interface{}) {
		value := fn()
		ReplaceAnchorMap[fmt.Sprintf("%s%s",
			rpsType, AnchorFlag)] = fmt.Sprintf("%s\n\r%s", AnchorFlag, value)
	}
)

// ProjectInfo
type ProjectInfo struct {
	ProjectName   string `yaml:"project_name"`
	ProjectPath   string `yaml:"project_path"`
	ProjectModule string `yaml:"project_module"`
}

// ReplaceKeywords 标识字典
type ReplaceKeywords struct {
	APP_NAME          string
	MODULE_NAME       string
	SERVICE_NAME      string
	COMPONENT_NAME    string
	COMPONENT_TYPE    string
	COMPONENT_UP_TYPE string
}
