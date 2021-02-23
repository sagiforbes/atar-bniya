package shell

import (
	"os"
	"strings"

	"github.com/robertkrimen/otto"
	"github.com/sagiforbes/banai/utils/ottoutils"
	"github.com/sagiforbes/banai/utils/shellutils"
	"github.com/sagiforbes/banai/utils/sshutils"
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

func shell(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		//		var e error
		var v otto.Value
		//		var opt shellutils.CommandOptions

		if len(call.ArgumentList) < 1 {
			return otto.Value{}
		}

		v = call.Otto.MakeCustomError("name", "message from error")

		return v

		// var cmd string
		// if len(call.ArgumentList) > 0 {
		// 	cmd = call.ArgumentList[0].String()
		// }

		// var ret *shellutils.ShellResult
		// if len(call.ArgumentList) > 1 {
		// 	var v = call.ArgumentList[1]
		// 	if v.IsObject() {
		// 		e = ottoutils.Val2Struct(v, &opt)
		// 		if e != nil {
		// 			logger.Panic("Failed to get shell options", e)

		// 		}

		// 	}
		// 	ret, e = shellutils.RunShellCommand(cmd, opt)
		// } else {
		// 	ret, e = shellutils.RunShellCommand(cmd)
		// }
		// if e != nil {
		// 	logger.Panicf("Fiailed to execute command in shell: %s", e)
		// }
		// v, e = vm.ToValue(ret)
		// if e != nil {
		// 	logger.Error("Failed to run shell", e)
		// 	panic(e)
		// }
		// return v
	}

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
		var sshConf shellutils.ShellSSHConfig
		if e = ottoutils.Val2Struct(v, &sshConf); e != nil {
			logger.Panic("Invalid ssh option variable")
		}

		var cmd string
		v = call.ArgumentList[1]
		if !v.IsString() {
			logger.Panic("Command must be a string")

		}
		cmd = v.String()

		var ret *shellutils.ShellResult

		ret, e = shellutils.RunRemoteShell(sshConf, cmd)
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
		var sshConf shellutils.ShellSSHConfig
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
		var sshConf shellutils.ShellSSHConfig
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
