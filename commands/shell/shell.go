package shell

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/robertkrimen/otto"
	"github.com/sagiforbes/banai/infra"
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

func panicOnError(e error, t ...string) {
	if e != nil {
		if t != nil {
			logger.Panic(t)
			logger.Panic(e)
		} else {
			logger.Panic(e)
		}

	}
}

func environments(b *infra.Banai) otto.Value {

	v, e := b.Jse.ToValue(envToMap())
	if e != nil {
		logger.Panic("Failed to translate env vars:")

	}
	return v
}

func callShell(b *infra.Banai, call otto.FunctionCall) (otto.Value, error) {
	var e error

	var opt shellutils.CommandOptions

	if len(call.ArgumentList) < 1 {
		return otto.Value{}, nil
	}

	var cmd string
	if len(call.ArgumentList) > 0 {
		cmd = call.ArgumentList[0].String()
	}

	var ret *shellutils.ShellResult
	if len(call.ArgumentList) > 1 {
		var v = call.ArgumentList[1]
		if v.IsObject() {
			e = ottoutils.Val2Struct(v, &opt)
			if e != nil {
				logger.Panic("Failed to get shell options", e)

			}

		}
		if opt.SecretID != "" {
			secret, err := b.GetSecret(opt.SecretID)
			panicOnError(err)
			switch secret.GetType() {
			case "ssh":
				s := secret.(infra.SSHWithPrivate)
				if opt.Env == nil {
					opt.Env = make([]string, 0)
				}
				opt.Env = append(opt.Env, fmt.Sprintf(`GIT_SSH_COMMAND="ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no"`, s.PrivatekeyFile))

			}

		}
		ret, e = shellutils.RunShellCommand(cmd, opt)
	} else {
		ret, e = shellutils.RunShellCommand(cmd)
	}
	if e != nil {
		logger.Panicf("Fiailed to execute command in shell: %s", e)
	}
	return call.Otto.ToValue(ret)
}

func shellScript(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) < 1 {
			return otto.Value{}
		}
		fileName := call.ArgumentList[0].String()
		fileContent, e := ioutil.ReadFile(fileName)
		panicOnError(e)
		v, _ := call.Otto.ToValue(string(fileContent))
		call.ArgumentList[0] = v
		v, e = callShell(b, call)
		panicOnError(e)
		return v
	}
}

func shell(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		v, e := callShell(b, call)
		panicOnError(e)
		return v
	}

}

func updateSSHConfigBySecret(b *infra.Banai, secretID string, sshConf *shellutils.ShellSSHConfig) {
	if sshConf.SecretID != "" {
		v, err := b.GetSecret(sshConf.SecretID)
		if err != nil {
			panicOnError(err)
		}
		switch v.GetType() {
		case "ssh":
			s := v.(infra.SSHWithPrivate)
			sshConf.Passphrase = s.Passfrase
			sshConf.PrivateKeyFile = s.PrivatekeyFile
			sshConf.User = s.User
		case "userpass":
			s := v.(infra.UserPassword)
			sshConf.User = s.User
			sshConf.Password = s.Password
		}

	}

}

func remoteshell(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
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
			panicOnError(e, "Invalid ssh option variable")
		}

		updateSSHConfigBySecret(b, sshConf.SecretID, &sshConf)
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
		v, e = call.Otto.ToValue(ret)
		if e != nil {
			logger.Panic("Failed to run shell", e)
		}
		return v
	}
}

func uploadFile(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
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

		updateSSHConfigBySecret(b, sshConf.SecretID, &sshConf)

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

func downloadFile(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
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
		updateSSHConfigBySecret(b, sshConf.SecretID, &sshConf)

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

func currentPath(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		s, err := os.Getwd()
		if err == nil {
			v, _ := call.Otto.ToValue(s)
			return v
		}
		logger.Panic("Failed to get current working folder ", err)
		return otto.Value{}
	}
}

func changeDir(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
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

//RegisterJSObjects registers Shell objects and functions
func RegisterJSObjects(b *infra.Banai) {
	logger = b.Logger
	b.Jse.Set("env", environments(b))
	b.Jse.Set("pwd", currentPath(b))
	b.Jse.Set("cd", changeDir(b))
	b.Jse.Set("sh", shell(b))
	b.Jse.Set("shScript", shellScript(b))
	b.Jse.Set("rsh", remoteshell(b))
	b.Jse.Set("shUpload", uploadFile(b))
	b.Jse.Set("shDownload", downloadFile(b))
}

func exampleImplementation(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		return otto.Value{}
	}
}
