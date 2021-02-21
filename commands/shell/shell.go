package shell

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/sagiforbes/banai/ottoutils"
	"github.com/sagiforbes/banai/sshutils"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var logger *logrus.Logger

func envToMap() map[string]string {
	var asMap = make(map[string]string)
	var eqIdx int
	for _, val := range os.Environ() {
		eqIdx = strings.IndexRune(val, '=')
		if eqIdx < 0 {
			asMap[strings.TrimSpace(val)] = "1"
		} else {
			if eqIdx == len(val)-1 {
				asMap[strings.TrimSpace(val[0:eqIdx])] = "1"
			} else {
				asMap[strings.TrimSpace(val[0:eqIdx])] = strings.TrimSpace(val[eqIdx+1:])
			}

		}
	}
	return asMap
}

func environments(vm *otto.Otto) otto.Value {

	v, e := vm.ToValue(envToMap())
	if e != nil {
		logger.Panic("Failed to translate env vars:")

	}
	return v
}

type shellResult struct {
	Code int    `json:"code,omitempty"`
	Out  string `json:"out,omitempty"`
	Err  string `json:"err,omitempty"`
}
type commandOptions struct {
	Shell   string   `json:"shell,omitempty"`
	In      string   `json:"in,omitempty"`
	Ins     []string `json:"ins,omitempty"`
	Timeout int      `json:"timeout,omitempty"`
}

func runShellCommand(commandToRun string, cmdOpt ...commandOptions) (ret shellResult) {

	var finalCommand = make([]string, 0)

	var shellCmd = "/bin/bash"
	if cmdOpt != nil {
		if cmdOpt[0].Shell != "" {
			shellCmd = cmdOpt[0].Shell
		}

	}

	finalCommand = append(finalCommand, "-c")
	finalCommand = append(finalCommand, commandToRun)

	logger.Info("Running shell script:", shellCmd, strings.Join(finalCommand, " "))
	var command *exec.Cmd
	if cmdOpt != nil && cmdOpt[0].Timeout > 0 {
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(cmdOpt[0].Timeout)*time.Second)
		defer cancelFunc()
		command = exec.CommandContext(ctx, shellCmd, finalCommand...)
	} else {
		command = exec.Command(shellCmd, finalCommand...)
	}

	command.Env = append(command.Env, os.Environ()...)

	cmdStdOutPipe, _ := command.StdoutPipe()
	cmdStdErrPipe, _ := command.StderrPipe()
	procWriter, _ := command.StdinPipe()

	var err error
	err = command.Start()
	if err != nil {
		logger.Panic("Failed to start command", err)
	}

	if cmdOpt != nil && cmdOpt[0].In != "" {
		procWriter.Write([]byte(cmdOpt[0].In))
		procWriter.Write([]byte("\n"))
	}
	if cmdOpt != nil && cmdOpt[0].Ins != nil {
		for _, line := range cmdOpt[0].Ins {
			procWriter.Write([]byte(line))
			procWriter.Write([]byte("\n"))
		}
	}
	var buf []byte
	buf, err = ioutil.ReadAll(cmdStdOutPipe)
	if err == nil {
		ret.Out = string(buf)
	}
	buf, err = ioutil.ReadAll(cmdStdErrPipe)
	if err == nil {
		ret.Err = string(buf)
	}

	err = command.Wait()
	if err == nil {
		ret.Code = 0
	} else {
		if exiterr, ok := err.(*exec.ExitError); ok {
			ret.Code = exiterr.ExitCode()
		} else {
			logger.Warn("Shell exit with error", err)
			ret.Code = 1
		}

	}

	return
}

func shell(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		var e error
		var v otto.Value
		var opt commandOptions

		var cmd string
		cmd = call.ArgumentList[0].String()

		var ret shellResult
		if len(call.ArgumentList) > 1 {
			var v = call.ArgumentList[1]
			if v.IsObject() {
				e = ottoutils.Val2Struct(v, &opt)
				if e != nil {
					logger.Panic("Failed to get shell options", e)

				}

			}
			ret = runShellCommand(cmd, opt)
		} else {
			ret = runShellCommand(cmd)
		}

		v, e = vm.ToValue(ret)
		if e != nil {
			logger.Error("Failed to run shell", e)
			panic(e)
		}
		return v
	}

}

type sshConfig struct {
	Address        string `json:"address,omitempty"`
	User           string `json:"user,omitempty"`
	Password       string `json:"password,omitempty"`
	PrivateKeyFile string `json:"privateKeyFile,omitempty"`
	Passphrase     string `json:"passphrase,omitempty"`
}

func runRemoteShell(sshConf sshConfig, cmd string) (*shellResult, error) {
	var sshClientConf *ssh.ClientConfig
	var e error

	if sshConf.Address == "" {
		return nil, fmt.Errorf("sshConfig target host Address not set")
	}
	if sshConf.User == "" {
		return nil, fmt.Errorf("sshConfig User not set")
	}

	if sshConf.Password != "" {
		sshClientConf = sshutils.CreateFromUserPassword(sshConf.User, sshConf.Password)
	} else {
		if sshConf.PrivateKeyFile != "" {
			sshClientConf, e = sshutils.CreateFromPrivateKeyFile(sshConf.User, sshConf.PrivateKeyFile, sshConf.Passphrase)
			if e != nil {
				return nil, e
			}
		}
	}

	var client *sshutils.Client
	client, e = sshutils.Dial(sshConf.Address, sshClientConf)
	if e != nil {
		return nil, e
	}
	defer client.Close()

	logger.Info("Running on remote machine: ", sshConf.Address, " ", cmd)
	stdout, stderr, e := client.Cmd(cmd)
	if e != nil {
		return nil, e
	}

	if len(stderr) > 0 {
		return &shellResult{
			Code: 1,
			Out:  string(stdout),
		}, nil
	}

	return &shellResult{
		Code: 0,
		Out:  string(stdout),
	}, nil
}

func remoteshell(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		var e error
		var v otto.Value

		if len(call.ArgumentList) != 2 {
			logger.Panic("Function must have two parameters, sshOpt and cmd as string")
		}

		v = call.ArgumentList[0]
		if !v.IsObject() {
			logger.Panic("Invalid ssh option variable")
		}
		var sshConf sshConfig
		if e = ottoutils.Val2Struct(v, &sshConf); e != nil {
			logger.Panic("Invalid ssh option variable")
		}

		var cmd string
		v = call.ArgumentList[1]
		if !v.IsString() {
			logger.Panic("Command must be a string")

		}
		cmd = v.String()

		var ret *shellResult

		ret, e = runRemoteShell(sshConf, cmd)
		if e != nil {
			logger.Panic("Failed to execute remote shell: ", e)

		}
		v, e = vm.ToValue(ret)
		if e != nil {
			logger.Panic("Failed to run shell", e)
		}
		return v
	}
}

func uploadFile(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		var e error
		var v otto.Value

		if len(call.ArgumentList) != 3 {
			logger.Panic("Function parameters are sshOpt local file and remote file")
		}

		v = call.ArgumentList[0]
		if !v.IsObject() {
			logger.Panic("Invalid ssh option variable")
		}
		var sshConf sshConfig
		if e = ottoutils.Val2Struct(v, &sshConf); e != nil {
			logger.Panic("Invalid ssh option variable")
		}
		var localFile string
		v = call.ArgumentList[1]
		if !v.IsString() {
			logger.Panic("Local file not set")

		}
		localFile = v.String()

		if stat, e := os.Stat(localFile); e != nil {
			logger.Panic("Failed to open local file", e)
		} else {
			if stat.IsDir() {
				logger.Panic("Local file is directory", e)
			}
		}

		var remoteFile string
		v = call.ArgumentList[2]
		if !v.IsString() {
			logger.Panic("Local file not set")

		}
		remoteFile = v.String()

		var sshClientConf *ssh.ClientConfig

		if sshConf.Address == "" {
			logger.Panic("sshConfig target host Address not set")
		}
		if sshConf.User == "" {
			logger.Panic("sshConfig User not set")
		}

		if sshConf.Password != "" {
			sshClientConf = sshutils.CreateFromUserPassword(sshConf.User, sshConf.Password)
		} else {
			if sshConf.PrivateKeyFile != "" {
				sshClientConf, e = sshutils.CreateFromPrivateKeyFile(sshConf.User, sshConf.PrivateKeyFile, sshConf.Passphrase)
				if e != nil {
					logger.Panic(e)
				}
			}
		}

		var client *sshutils.Client
		client, e = sshutils.Dial(sshConf.Address, sshClientConf)
		if e != nil {
			logger.Panic(e)
		}

		defer client.Close()

		e = client.UploadFile(localFile, remoteFile)
		if e != nil {
			logger.Panic(e)
		}

		return otto.Value{}
	}
}

func downloadFile(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		var e error
		var v otto.Value

		if len(call.ArgumentList) != 3 {
			logger.Panic("Function parameters are sshOpt local file and remote file")
		}

		v = call.ArgumentList[0]
		if !v.IsObject() {
			logger.Panic("Invalid ssh option variable")
		}
		var sshConf sshConfig
		if e = ottoutils.Val2Struct(v, &sshConf); e != nil {
			logger.Panic("Invalid ssh option variable")
		}
		var remoteFile string
		v = call.ArgumentList[1]
		if !v.IsString() {
			logger.Panic("Local file not set")

		}
		remoteFile = v.String()

		if stat, e := os.Stat(remoteFile); e != nil {
			logger.Panic("Failed to open local file", e)
		} else {
			if stat.IsDir() {
				logger.Panic("Local file is directory", e)
			}
		}

		var localFile string
		v = call.ArgumentList[2]
		if !v.IsString() {
			logger.Panic("Local file not set")

		}
		localFile = v.String()

		var sshClientConf *ssh.ClientConfig

		if sshConf.Address == "" {
			logger.Panic("sshConfig target host Address not set")
		}
		if sshConf.User == "" {
			logger.Panic("sshConfig User not set")
		}

		if sshConf.Password != "" {
			sshClientConf = sshutils.CreateFromUserPassword(sshConf.User, sshConf.Password)
		} else {
			if sshConf.PrivateKeyFile != "" {
				sshClientConf, e = sshutils.CreateFromPrivateKeyFile(sshConf.User, sshConf.PrivateKeyFile, sshConf.Passphrase)
				if e != nil {
					logger.Panic(e)
				}
			}
		}

		var client *sshutils.Client
		client, e = sshutils.Dial(sshConf.Address, sshClientConf)
		if e != nil {
			logger.Panic(e)
		}

		defer client.Close()

		e = client.Download(remoteFile, localFile)
		if e != nil {
			logger.Panic(e)
		}

		return otto.Value{}
	}
}

func currentPath(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		s, err := os.Getwd()
		if err == nil {
			v, _ := vm.ToValue(s)
			return v
		}
		logger.Panic("Failed to get current working folder ", err)
		return otto.Value{}
	}
}

func changeDir(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 1 {
			return otto.Value{}
		}
		err := os.Chdir(call.ArgumentList[0].String())
		if err != nil {
			logger.Panic("Failed to get current working folder ", err)
		}

		return otto.Value{}
	}
}

//RegisterObjects registers Shell objects and functions
func RegisterObjects(vm *otto.Otto, lgr *logrus.Logger) {
	logger = lgr
	vm.Set("env", environments(vm))
	vm.Set("pwd", currentPath(vm))
	vm.Set("cd", changeDir(vm))
	vm.Set("sh", shell(vm))
	vm.Set("rsh", remoteshell(vm))
	vm.Set("shUpload", uploadFile(vm))
	vm.Set("shDownload", downloadFile(vm))
}

func exampleImplementation(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		return otto.Value{}
	}
}
