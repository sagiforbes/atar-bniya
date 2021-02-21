package docker

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"os"

	"github.com/robertkrimen/otto"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func genericHashCalculator(hasher hash.Hash, src io.Reader) string {
	buf := make([]byte, 1024)
	io.CopyBuffer(hasher, src, buf)
	return hex.EncodeToString(hasher.Sum(nil))
}

func getFileFromCaller(call otto.FunctionCall) *os.File {
	if len(call.ArgumentList) != 1 {
		logger.Panic("No file to calc")
	}
	fileName := call.ArgumentList[0].String()
	fi, err := os.Stat(fileName)
	if err != nil {
		logger.Panicf("Cannot calculate hash for %s %s", fileName, err)
	}
	if !fi.Mode().IsRegular() {
		logger.Panicf("Cannot calculate hash for %s", fileName)
	}
	f, err := os.Open(fileName)
	if err != nil {
		logger.Panicf("Failed to open file %s, for reading: %e", fileName, err)
	}

	return f
}

//********************* MD5 *************************
func hashMD5Buf(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 1 {
			logger.Panic("No buffer to calculate")
		}
		val, _ := call.ArgumentList[0].Export()
		bs, ok := val.([]byte)
		if !ok {
			logger.Panic("No buffer to calculate")
		}
		b := bytes.NewBuffer(bs)
		v, _ := vm.ToValue(genericHashCalculator(md5.New(), b))
		return v
	}
}

func hashMD5Text(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 1 {
			logger.Panic("No String to calculate")
		}

		b := bytes.NewBufferString(call.ArgumentList[0].String())
		v, _ := vm.ToValue(genericHashCalculator(md5.New(), b))
		return v
	}
}

func hashMD5File(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		f := getFileFromCaller(call)
		defer f.Close()

		v, _ := vm.ToValue(genericHashCalculator(md5.New(), f))

		return v
	}
}

//********************* SHA1 *************************
func sha1Buf(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 1 {
			logger.Panic("No buffer to calculate")
		}
		val, _ := call.ArgumentList[0].Export()
		bs, ok := val.([]byte)
		if !ok {
			logger.Panic("No buffer to calculate")
		}
		b := bytes.NewBuffer(bs)
		v, _ := vm.ToValue(genericHashCalculator(sha1.New(), b))
		return v
	}
}

func sha1Text(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 1 {
			logger.Panic("No String to calculate")
		}

		b := bytes.NewBufferString(call.ArgumentList[0].String())
		v, _ := vm.ToValue(genericHashCalculator(sha1.New(), b))
		return v
	}
}

func sha1File(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		f := getFileFromCaller(call)
		defer f.Close()

		v, _ := vm.ToValue(genericHashCalculator(sha1.New(), f))

		return v
	}
}

//********************* SHA256 *************************
func sha256Buf(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 1 {
			logger.Panic("No buffer to calculate")
		}
		val, _ := call.ArgumentList[0].Export()
		bs, ok := val.([]byte)
		if !ok {
			logger.Panic("No buffer to calculate")
		}
		b := bytes.NewBuffer(bs)
		v, _ := vm.ToValue(genericHashCalculator(sha256.New(), b))
		return v
	}
}

func sha256Text(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 1 {
			logger.Panic("No String to calculate")
		}

		b := bytes.NewBufferString(call.ArgumentList[0].String())
		v, _ := vm.ToValue(genericHashCalculator(sha256.New(), b))
		return v
	}
}

func sha256File(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		f := getFileFromCaller(call)
		defer f.Close()

		v, _ := vm.ToValue(genericHashCalculator(sha256.New(), f))

		return v
	}
}

//RegisterObjects registers Shell objects and functions
func RegisterObjects(vm *otto.Otto, lgr *logrus.Logger) {
	logger = lgr
	vm.Set("hashMd5File", hashMD5File(vm))
	vm.Set("hashMd5Text", hashMD5Text(vm))
	vm.Set("hashMd5Buffer", hashMD5Buf(vm))
	vm.Set("hashSha1File", sha1File(vm))
	vm.Set("hashSha1Text", sha1Text(vm))
	vm.Set("hashSha1Buffer", sha1Buf(vm))
	vm.Set("hashSha256File", sha256File(vm))
	vm.Set("hashSha256Text", sha256Text(vm))
	vm.Set("hashSha256Buffer", sha256Buf(vm))

}

func exampleImplementation(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		return otto.Value{}
	}
}
