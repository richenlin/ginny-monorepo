package handler

import (
	"errors"
	"fmt"

	"github.com/fatih/structs"
	"github.com/goriller/ginny-cli/ginny/options"
	"github.com/goriller/ginny-cli/ginny/util"
	"github.com/iancoleman/strcase"
)

func CreateComponent(repoName string, comType string) error {
	conf, err := GetProjectInfo()
	if err != nil {
		util.Error("Failed to get project info, ", err.Error())
		return err
	}

	tmpPath := GetTempPath(options.TempPath)
	if err := PullTemplate(tmpPath, options.ComponentTemplateRepo); err != nil {
		return err
	}

	srcFile := fmt.Sprintf("%s/component/tpl.go", tmpPath)
	dstFile := fmt.Sprintf("%s/internal/%s/%s.go", conf.ProjectPath, comType, repoName)
	if util.Exists(dstFile) {
		return errors.New("File already exists and overwriting is not allowed")
	}
	if err := util.CopyFile(srcFile, dstFile); err != nil {
		return err
	}
	// replace
	camelName := strcase.ToCamel(repoName)
	camelType := strcase.ToCamel(comType)
	r := &options.ReplaceKeywords{
		APP_NAME:          conf.ProjectName,
		MODULE_NAME:       conf.ProjectModule,
		COMPONENT_NAME:    camelName,
		COMPONENT_TYPE:    comType,
		COMPONENT_UP_TYPE: camelType,
	}

	// replace provider
	providerFile := fmt.Sprintf("%s/internal/%s/provider.go", conf.ProjectPath, comType)
	if !util.Exists(providerFile) {
		if err := util.CopyFile(fmt.Sprintf("%s/component/provider.go", tmpPath), providerFile); err != nil {
			return err
		}
	}
	m := structs.Map(r)
	for k, v := range m {
		options.ReplaceAnchorMap[k] = v
	}

	options.RegReplaceWithAnchor("//COMPONENT_PROVIDER", func() interface{} {
		//COMPONENT_NAMECOMPONENT_UP_TYPEProvider
		return fmt.Sprintf("%s%sProviderSet", camelName, camelType)
	})

	if err := ReplaceFileKeyword([]string{dstFile, providerFile}, options.ReplaceAnchorMap); err != nil {
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
