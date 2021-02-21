package fs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/robertkrimen/otto"
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

func readBinFile(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 1 {
			logger.Panic("Must have name of file")
		}
		v := call.ArgumentList[0]
		if !v.IsString() {
			logger.Panic("Invalid file name")
		}
		fileName := v.String()
		b, err := readFileContent(fileName)
		if err != nil {
			logger.Panic("Error reading", fileName, err)
		}

		v, err = vm.ToValue(b)
		if err != nil {
			logger.Panic(err)
		}
		return v
	}
}

func readFile(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 1 {
			logger.Panic("Must have name of file")
		}
		v := call.ArgumentList[0]
		if !v.IsString() {
			logger.Panic("Invalid file name")
		}
		fileName := v.String()
		b, err := readFileContent(fileName)
		if err != nil {
			logger.Panic("Error reading", fileName, err)
		}

		v, err = vm.ToValue(string(b))
		if err != nil {
			logger.Panic(err)
		}
		return v
	}
}

func writeFile(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 2 {
			logger.Panic("Must have file name and file content as text")
		}
		v := call.ArgumentList[0]
		if !v.IsString() {
			logger.Panic("Invalid file name")
		}
		fileName := v.String()

		v = call.ArgumentList[1]
		if !v.IsString() {
			logger.Panic("File content must be as text")
		}
		content := v.String()

		err := ioutil.WriteFile(fileName, []byte(content), 0644)
		if err != nil {
			logger.Panic("Error writing file:", fileName, err)
		}

		return otto.Value{}
	}
}

func writeBinFile(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 2 {
			logger.Panic("Must have file name and file content as array of bytes")
		}
		v := call.ArgumentList[0]
		if !v.IsString() {
			logger.Panic("Invalid file name")
		}
		fileName := v.String()

		v = call.ArgumentList[1]
		if !v.IsObject() {
			logger.Panic("File content must be as array of bytes")
		}
		content, err := v.Export()
		if err != nil {
			logger.Panic("Failed to read file content")
		}

		err = ioutil.WriteFile(fileName, content.([]byte), 0644)
		if err != nil {
			logger.Panic("Error writing file:", fileName, err)
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

func fsRemove(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) < 1 {
			logger.Panic("Name of element to remove not set")
		}
		v := call.ArgumentList[0]
		if !v.IsString() {
			logger.Panic("Invalid element name")
		}
		itemName := v.String()
		_, err := os.Stat(itemName)
		if os.IsNotExist(err) {
			return otto.Value{}
		}

		deleteAll := false
		if len(call.ArgumentList) > 1 {
			v := call.ArgumentList[1]
			if v.IsBoolean() {
				deleteAll, _ = v.ToBoolean()
			}
		}
		if deleteAll {
			logger.Info("Deleting all under ", itemName)
			err = os.RemoveAll(itemName)
			if err != nil {
				if !os.IsNotExist(err) {
					logger.Panic("Failed to delete ", itemName, err)
				}
			}
		} else {
			err = os.Remove(itemName)
			if err != nil {
				if !os.IsNotExist(err) {
					logger.Panic("Failed to delete ", itemName, err)
				}
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

func fileCopy(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 2 {
			logger.Panic("copy must have two parameters, source file and destination file")
		}
		v := call.ArgumentList[0]
		if !v.IsString() {
			logger.Panic("Invalid source file name")
		}

		sourceFileName := v.String()

		v = call.ArgumentList[1]
		if !v.IsString() {
			logger.Panic("Invalid destination file name")
		}
		destinationFileName := v.String()

		err := copyFiles(sourceFileName, destinationFileName)
		if err != nil {
			logger.Panic("Failed to copy files ", err)
		}

		return otto.Value{}
	}
}

func fileMove(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 2 {
			logger.Panic("copy must have two parameters, source file and destination file")
		}
		v := call.ArgumentList[0]
		if !v.IsString() {
			logger.Panic("Invalid source file name")
		}

		sourceFileName := v.String()

		v = call.ArgumentList[1]
		if !v.IsString() {
			logger.Panic("Invalid destination file name")
		}
		destinationFileName := v.String()

		err := copyFiles(sourceFileName, destinationFileName)
		if err != nil {
			logger.Panic("Failed to move files ", err)
		}
		err = os.Remove(sourceFileName)
		if err != nil {
			logger.Panicf("Failed to move source file %s, %s", sourceFileName, err)
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
			v = call.ArgumentList[0]
			if v.IsString() {
				rootFolder = v.String()
			}
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
	vm.Set("fsReadBin", readBinFile(vm))
	vm.Set("fsWrite", writeFile(vm))
	vm.Set("fsWriteBin", writeBinFile(vm))
	vm.Set("fsCreateDir", createDir(vm))
	vm.Set("fsRemoveDir", fsRemove(vm))
	vm.Set("fsRemoveFile", fsRemove(vm))
	vm.Set("fsRemove", fsRemove(vm))
	vm.Set("fsCopy", fileCopy(vm))
	vm.Set("fsMove", fileMove(vm))
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
