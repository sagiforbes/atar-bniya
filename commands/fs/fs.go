package fs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/robertkrimen/otto"
	"github.com/sagiforbes/banai/utils/fsutils"
	"github.com/sagiforbes/banai/utils/shellutils"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func readFileContent(fileName string) ([]byte, error) {
	info, err := os.Stat(fileName)
	if err != nil {
		logger.Panic("Error reading", fileName, err)
	}
	if info.IsDir() {
		logger.Panic("file", fileName, "Is a directory")
	}
	return ioutil.ReadFile(fileName)

}

func readFile(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 1 {
			logger.Panic("Must have name of file")
		}

		fileName := call.ArgumentList[0].String()
		b, err := readFileContent(fileName)
		if err != nil {
			logger.Panic("Error reading", fileName, err)
		}

		var v otto.Value
		if !utf8.Valid(b) {
			v, err = vm.ToValue(b)
			if err != nil {
				logger.Panic(err)
			}
		} else {
			v, err = vm.ToValue(string(b))
			if err != nil {
				logger.Panic(err)
			}
		}

		return v
	}
}

func writeFile(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 2 {
			logger.Panic("Must have file name and file content")
		}
		fileName := call.ArgumentList[0].String()

		v := call.ArgumentList[1]
		content := make([]byte, 0)
		if v.IsString() {
			content = []byte(call.ArgumentList[1].String())
		} else {
			arbitrary, err := v.Export()
			if err != nil {
				logger.Panicf("Cannot save this information to file, %s", err)
			}
			byteArray, ok := arbitrary.([]uint8)
			if !ok {
				logger.Panic("Cannot save this information to file. Content should be string or bytes")
			}
			content = byteArray
		}

		err := ioutil.WriteFile(fileName, content, 0644)
		if err != nil {
			logger.Panicf("Error writing file %s: %s", fileName, err)
		}

		return otto.Value{}
	}
}

func createDir(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 1 {
			logger.Panic("Name of dir not set")
		}
		v := call.ArgumentList[0]
		if !v.IsString() {
			logger.Panic("Invalid file name")
		}
		dirName := v.String()
		s, err := os.Stat(dirName)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dirName, 0755); err != nil {
				logger.Panic("Failed to create dir ", err)
			}
		} else {
			if !s.IsDir() {
				logger.Panic("Already have file by that name")
			}
		}

		return otto.Value{}
	}
}

func fsRemoveDir(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) < 1 {
			logger.Panic("Name of element to remove not set")
		}
		itemName := call.ArgumentList[0].String()
		logger.Info("Deleting all under ", itemName)
		err := os.RemoveAll(itemName)
		if err != nil {
			if !os.IsNotExist(err) {
				logger.Panicf("Failed to delete %s, %s", itemName, err)
			}
		}
		return otto.Value{}
	}
}

func fsRemove(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) < 1 {
			logger.Panic("Name of element to remove not set")
		}
		itemName := call.ArgumentList[0].String()

		err := os.Remove(itemName)
		if err != nil {
			if !os.IsNotExist(err) {
				logger.Panicf("Failed to delete %s, %s", itemName, err)
			}
		}

		return otto.Value{}
	}
}

func copyFiles(sourceFileName, destinationFileName string) error {
	fileState, err := os.Stat(sourceFileName)
	if err != nil {
		return err
	}
	if !fileState.Mode().IsRegular() {
		return fmt.Errorf("Invalid source file")
	}
	fileState, err = os.Stat(destinationFileName)
	if err == nil {
		if fileState.IsDir() {
			destinationFileName = filepath.Join(destinationFileName, sourceFileName)
		}
	}

	logger.Infof("Copy %s -> %s", sourceFileName, destinationFileName)

	source, err := os.Open(sourceFileName)
	if err != nil {
		return fmt.Errorf("Failed to open source file %s for copy %s", sourceFileName, err)
	}
	defer source.Close()

	destination, err := os.OpenFile(destinationFileName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("Failed to open destination file %s for copy %s", destinationFileName, err)
	}
	defer destination.Close()

	copyBuf := make([]byte, 4096)
	_, err = io.CopyBuffer(destination, source, copyBuf)
	return err
}

func fsCopy(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 2 {
			logger.Panic("copy must have two parameters, source and destination")
		}
		sourceFileName := call.ArgumentList[0].String()

		destinationFileName := call.ArgumentList[1].String()

		err := fsutils.CopyfsItem(sourceFileName, destinationFileName)
		if err != nil {
			logger.Panic("Failed to copy files ", err)
		}

		return otto.Value{}
	}
}

func fsMove(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 2 {
			logger.Panic("Move must have two parameters, source and destination")
		}

		sourceFileName := call.ArgumentList[0].String()
		destinationFileName := call.ArgumentList[1].String()

		result, err := shellutils.RunShellCommand(fmt.Sprintf("mv %s %s", sourceFileName, destinationFileName))
		if err != nil {
			logger.Panic("Failed to move ", err)
		}
		if result.Code != 0 {
			logger.Panic("Move exit code ", result.Code)
		}

		return otto.Value{}
	}
}

type splitPathNameParts struct {
	Folder string `json:"folder,omitempty"`
	File   string `json:"file,omitempty"`
	Title  string `json:"title,omitempty"`
	Ext    string `json:"ext,omitempty"`
}

func splitPathNameComponents(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		var v otto.Value
		if len(call.ArgumentList) != 1 {
			v, _ = vm.ToValue(splitPathNameParts{})
			return v
		}
		v = call.ArgumentList[0]
		if !v.IsString() {
			logger.Panic("Invalid path name")
		}
		pathName, _ := v.ToString()

		dirName, fileName := filepath.Split(pathName)
		extName := filepath.Ext(fileName)

		ret := splitPathNameParts{
			Folder: filepath.Clean(dirName),
			File:   fileName,
			Ext:    extName,
		}
		if ret.Ext != "" {
			ret.Title = ret.File[:len(ret.File)-len(ret.Ext)]
		} else {
			ret.Title = ret.File
		}

		v, _ = vm.ToValue(ret)

		return v
	}
}

func joinPathParts(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		var v otto.Value
		if len(call.ArgumentList) < 1 {
			v, _ = vm.ToValue("")
			return v
		}

		var parts = make([]string, len(call.ArgumentList))
		for i, arg := range call.ArgumentList {
			parts[i], _ = arg.ToString()
		}
		v, _ = vm.ToValue(filepath.Join(parts...))

		return v
	}
}

func listAllSubitemsInDir(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		var rootFolder = "."
		var itemType = ""
		var v otto.Value
		if len(call.ArgumentList) > 0 {
			rootFolder = call.ArgumentList[0].String()
		}

		if len(call.ArgumentList) > 1 {
			v = call.ArgumentList[1]
			if v.IsString() {
				itemType = v.String()
				switch itemType {
				case "f":
					itemType = "f"
				case "d":
					itemType = "d"
				default:
					itemType = ""
				}
			}
		}

		ret := make([]string, 0)
		err := filepath.Walk(rootFolder, func(path string, info os.FileInfo, err error) error {
			if err == nil {
				switch itemType {
				case "f":
					if info.Mode().IsRegular() {
						ret = append(ret, path)
					}
				case "d":
					if info.IsDir() {
						if path != rootFolder {
							ret = append(ret, path)
						}

					}
				default:
					ret = append(ret, path)
				}

			}
			return err
		})
		if err != nil {
			logger.Panicf("Failed to list folder %s, %s", rootFolder, err)
		}
		v, _ = vm.ToValue(ret)
		return v
	}
}

func absoluteFolder(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		var v otto.Value
		if len(call.ArgumentList) != 1 {
			v, _ = vm.ToValue("")
			return v
		}

		s, err := filepath.Abs(call.ArgumentList[0].String())
		if err != nil {
			logger.Panic("Failed to calculate absolute path: ", err)
		}
		v, _ = vm.ToValue(s)

		return v
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

//RegisterObjects registers fs objects and functions
func RegisterObjects(vm *otto.Otto, lgr *logrus.Logger) {
	logger = lgr
	vm.Set("fsRead", readFile(vm))
	vm.Set("fsWrite", writeFile(vm))
	vm.Set("fsCreateDir", createDir(vm))
	vm.Set("fsRemoveDir", fsRemoveDir(vm))
	vm.Set("fsRemove", fsRemove(vm))
	vm.Set("fsCopy", fsCopy(vm))
	vm.Set("fsMove", fsMove(vm))
	vm.Set("fsSplit", splitPathNameComponents(vm))
	vm.Set("fsJoin", joinPathParts(vm))
	vm.Set("fsList", listAllSubitemsInDir(vm))
	vm.Set("fsAbs", absoluteFolder(vm))
	vm.Set("fsPwd", currentPath(vm))
	vm.Set("fsChdir", changeDir(vm))
}

func exampleImplementation(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		return otto.Value{}
	}
}
