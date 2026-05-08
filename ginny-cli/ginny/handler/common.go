package handler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/emicklei/proto"
	"github.com/go-git/go-git/v5"
	"github.com/goriller/ginny-cli/ginny/options"
	"github.com/goriller/ginny-cli/ginny/util"
	"gopkg.in/yaml.v3"
)

// GetCurrentDir 获取当前目录
func GetCurrentDir() (string, error) {
	// 获取当前目录
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return dir, nil
}

// PullTemplate 拉取模板
func PullTemplate(dir, repo string) error {
	if !util.Exists(dir + "/.git") {
		// Clone the given repository to the given directory
		util.Info("git clone " + repo)
		_, err := git.PlainClone(dir, false, &git.CloneOptions{
			URL:      repo,
			Progress: os.Stdout,
		})
		if err != nil {
			util.Error("Clone the given repository error:", err.Error())
			return err
		}
		return nil
	}

	// We instantiate a new repository targeting the given path (the .git folder)
	r, err := git.PlainOpen(dir)
	if err != nil {
		return err
	}
	// Get the working directory for the repository
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	util.Info("git pull origin master ")
	err = w.Pull(&git.PullOptions{RemoteName: "origin", Force: true})
	if err != nil {
		util.Error(err.Error())
		return nil
	}
	// Print the latest commit that was just pulled
	ref, err := r.Head()
	if err != nil {
		util.Error(err.Error())
		return nil
	}
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		util.Error(err.Error())
		return nil
	}
	util.Info("git pull success", commit)

	return nil
}

// GenerateProjectInfo 构造项目标识
func GenerateProjectInfo(projectDir string, p *options.ProjectInfo) error {
	bt, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	return util.WriteToFile(projectDir+"/"+options.ProjectFlag, bt)
}

// GetProjectInfo 检查当前目录是否ginny项目，并返回项目信息
func GetProjectInfo() (*options.ProjectInfo, error) {
	dir, err := GetCurrentDir()
	if err != nil {
		return nil, errors.New("Failed to get project directory.")
	}

	flagFile := dir + "/" + options.ProjectFlag
	if !util.Exists(flagFile) {
		return nil, errors.New("Current project is not a Ginny project.\n" +
			"Please execute command after enter Ginny project root directory.")
	}

	bin, err := ioutil.ReadFile(flagFile)
	if err != nil {
		return nil, errors.New("Failed to read project flag file.")
	}
	conf := &options.ProjectInfo{}
	err = yaml.Unmarshal(bin, conf)
	if err != nil {
		return nil, errors.New("Failed Unmarshal projectinfo.")
	}
	if conf.ProjectName == "" {
		return nil, errors.New("The project flags file is corrupted .")
	}
	if conf.ProjectPath == "" {
		conf.ProjectPath = dir
	}

	return conf, nil
}

// ReplaceFileKeyword
func ReplaceFileKeyword(files []string, m map[string]interface{}) error {
	for _, f := range files {
		if !util.IsFile(f) {
			continue
		}
		for k, v := range m {
			if m[k] != "" {
				if err := util.ReplaceFile(f, k, fmt.Sprintf("%v", v)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ExecCommand
func ExecCommand(dir, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	util.Info("ExecCommand", cmd)
	err := cmd.Run()
	if err != nil {
		return err
	}
	util.Info("ExecCommand", "End")
	return nil
}

// GoFmtDir
func GoFmtDir(dir string) error {
	fs := util.GetFiles(dir)
	for _, v := range fs {
		if strings.HasSuffix(v, ".go") {
			if err := ExecCommand(dir, "gofmt", "-w", v); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetTempPath
func GetTempPath(path string) string {
	tmp := fmt.Sprintf("%s/%s", os.TempDir(), path)
	if !util.Exists(tmp) {
		_ = util.MkDir(tmp)
	}
	return tmp
}

// parseProtoFile 解析pb文件，拿到proto文件描述信息
func parseProtoFile(fname string) (map[string][]*proto.RPC, error) {
	reader, _ := os.Open(fname)
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	elem := map[string][]*proto.RPC{}
	proto.Walk(definition,
		proto.WithService(func(s *proto.Service) {
			elem[s.Name] = []*proto.RPC{}
			for _, each := range s.Elements {
				if c, ok := each.(*proto.RPC); ok {
					elem[s.Name] = append(elem[s.Name], c)
				}
			}
		}),
	)
	return elem, nil
}
