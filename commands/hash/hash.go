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
	"github.com/sagiforbes/banai/infra"
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
func hashMD5Buf(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
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
		v, _ := call.Otto.ToValue(genericHashCalculator(md5.New(), b))
		return v
	}
}

func hashMD5Text(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 1 {
			logger.Panic("No String to calculate")
		}

		b := bytes.NewBufferString(call.ArgumentList[0].String())
		v, _ := call.Otto.ToValue(genericHashCalculator(md5.New(), b))
		return v
	}
}

func hashMD5File(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		f := getFileFromCaller(call)
		defer f.Close()

		v, _ := call.Otto.ToValue(genericHashCalculator(md5.New(), f))

		return v
	}
}

//********************* SHA1 *************************
func sha1Buf(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
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
		v, _ := call.Otto.ToValue(genericHashCalculator(sha1.New(), b))
		return v
	}
}

func sha1Text(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 1 {
			logger.Panic("No String to calculate")
		}

		b := bytes.NewBufferString(call.ArgumentList[0].String())
		v, _ := call.Otto.ToValue(genericHashCalculator(sha1.New(), b))
		return v
	}
}

func sha1File(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		f := getFileFromCaller(call)
		defer f.Close()

		v, _ := call.Otto.ToValue(genericHashCalculator(sha1.New(), f))

		return v
	}
}

//********************* SHA256 *************************
func sha256Buf(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
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
		v, _ := call.Otto.ToValue(genericHashCalculator(sha256.New(), b))
		return v
	}
}

func sha256Text(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 1 {
			logger.Panic("No String to calculate")
		}

		b := bytes.NewBufferString(call.ArgumentList[0].String())
		v, _ := call.Otto.ToValue(genericHashCalculator(sha256.New(), b))
		return v
	}
}

func sha256File(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		f := getFileFromCaller(call)
		defer f.Close()

		v, _ := call.Otto.ToValue(genericHashCalculator(sha256.New(), f))

		return v
	}
}

//RegisterJSObjects registers Shell objects and functions
func RegisterJSObjects(b *infra.Banai) {
	logger = b.Logger
	b.Jse.Set("hashMd5File", hashMD5File(b))
	b.Jse.Set("hashMd5Text", hashMD5Text(b))
	b.Jse.Set("hashMd5Buffer", hashMD5Buf(b))
	b.Jse.Set("hashSha1File", sha1File(b))
	b.Jse.Set("hashSha1Text", sha1Text(b))
	b.Jse.Set("hashSha1Buffer", sha1Buf(b))
	b.Jse.Set("hashSha256File", sha256File(b))
	b.Jse.Set("hashSha256Text", sha256Text(b))
	b.Jse.Set("hashSha256Buffer", sha256Buf(b))

}

func exampleImplementation(b *infra.Banai) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		return otto.Value{}
	}
}
