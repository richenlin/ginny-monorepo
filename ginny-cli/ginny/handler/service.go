package handler

import (
	"fmt"
	"strings"

	"github.com/fatih/structs"
	"github.com/goriller/ginny-cli/ginny/options"
	"github.com/goriller/ginny-cli/ginny/util"
	"github.com/iancoleman/strcase"
)

func CreateService() error {
	conf, err := GetProjectInfo()
	if err != nil {
		util.Error("Failed to get project info, ", err.Error())
		return err
	}

	tmpPath := GetTempPath(options.TempPath)
	if err := PullTemplate(tmpPath, options.ComponentTemplateRepo); err != nil {
		return err
	}

	// 解析proto文件，自动生成 method 文件
	protoFile := fmt.Sprintf(options.ProtoPath, conf.ProjectPath)
	// protoPath := path.Dir(protoFile)
	fileDesc, err := parseProtoFile(protoFile)
	if err != nil {
		return err
	}
	var serviceName string
	replaceArr := []string{}
	for k, v := range fileDesc {
		serviceName = k
		for _, val := range v {
			fmt.Printf("%s: %v,%v,%v\n", k, val.Name, val.RequestType, val.ReturnsType)
			srcFile := fmt.Sprintf("%s/service/tpl.go", tmpPath)
			dstFile := fmt.Sprintf("%s/internal/service/%s.go", conf.ProjectPath, strings.ToLower(val.Name))
			if util.Exists(dstFile) {
				continue
			}
			if err := util.CopyFile(srcFile, dstFile); err != nil {
				return err
			}
			replaceArr = append(replaceArr, dstFile)
			if err := ReplaceFileKeyword([]string{dstFile}, map[string]interface{}{
				"METHOD_NAME":    val.Name,
				"METHOD_REQNAME": val.RequestType,
				"METHOD_RESNAME": val.ReturnsType,
			}); err != nil {
				return err
			}
		}
		//仅支持单个service
		break
	}

	// replace provider
	providerFile := fmt.Sprintf("%s/internal/service/provider.go", conf.ProjectPath)
	if !util.Exists(providerFile) {
		if err := util.CopyFile(fmt.Sprintf("%s/service/provider.go", tmpPath), providerFile); err != nil {
			return err
		}
		replaceArr = append(replaceArr, providerFile)
	}
	// replace service
	camelName := strcase.ToCamel(serviceName)
	r := &options.ReplaceKeywords{
		APP_NAME:     conf.ProjectName,
		MODULE_NAME:  conf.ProjectModule,
		SERVICE_NAME: camelName,
	}
	m := structs.Map(r)
	for k, v := range m {
		options.ReplaceAnchorMap[k] = v
	}

	if err := ReplaceFileKeyword(replaceArr, options.ReplaceAnchorMap); err != nil {
		return err
	}

	if err := ExecCommand(conf.ProjectPath, "go", "mod", "tidy"); err != nil {
		return err
	}
	if err := ExecCommand(conf.ProjectPath, "wire", "./cmd/."); err != nil {
		return err
	}
	_ = GoFmtDir(conf.ProjectPath)
	return nil
}
