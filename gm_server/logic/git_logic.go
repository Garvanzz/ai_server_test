package logic

import (
	"fmt"
	"os"
	"os/exec"
)

// GitResult 用于 Git 拉取/克隆结果
type GitResult struct {
	Success    bool
	Message    string
	Directory  string
	Branch     string
	LastCommit string
}

func GitPullOrClone(targetDir, repoURL string) GitResult {
	result := GitResult{Directory: targetDir}

	// 检查并切换到目标目录
	if err := os.Chdir(targetDir); err != nil {
		result.Message = fmt.Sprintf("无法进入目录: %v", err)
		return result
	}

	// 判断是否是Git仓库
	if isGitRepo() {
		return executeGitPull()
	}
	return executeGitClone(repoURL)
}

func isGitRepo() bool {
	_, err := os.Stat(".git")
	return !os.IsNotExist(err)
}

func executeGitPull() GitResult {
	result := GitResult{}

	resetCmd := exec.Command("git", "reset", "--hard", "origin/main")
	output, err := resetCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("git reset failed: %v\nOutput: %s\n", err, string(output))
		result.Message = fmt.Sprintf("拉取失败: %v\n%s", err, string(output))
		return result
	}

	cmd := exec.Command("git", "pull")
	output, err = cmd.CombinedOutput()

	if err != nil {
		result.Message = fmt.Sprintf("拉取失败: %v\n%s", err, string(output))
		return result
	}

	return getGitStatus("拉取成功")
}

func executeGitClone(repoURL string) GitResult {
	result := GitResult{}
	cmd := exec.Command("git", "clone", repoURL, ".")
	output, err := cmd.CombinedOutput()

	if err != nil {
		result.Message = fmt.Sprintf("克隆失败: %v\n%s", err, string(output))
		return result
	}

	return getGitStatus("克隆成功")
}

func getGitStatus(successMsg string) GitResult {
	result := GitResult{Success: true, Message: successMsg}

	// 获取当前目录绝对路径
	if dir, err := os.Getwd(); err == nil {
		result.Directory = dir
	}

	// 获取当前分支
	if branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		result.Branch = string(branch[:len(branch)-1]) // 去除换行符
	}

	// 获取最新提交
	if commit, err := exec.Command("git", "log", "-1", "--pretty=format:%h - %s (%cr)").Output(); err == nil {
		result.LastCommit = string(commit)
	}

	return result
}
