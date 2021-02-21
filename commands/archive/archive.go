package archive

import (
	"os"
	"path/filepath"

	"github.com/robertkrimen/otto"
	"github.com/sagiforbes/banai/fsutils"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func discoverFilesInFolder(folderPath string) []string {
	var files []string
	filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			if !info.IsDir() {
				files = append(files, path)
			}
			return nil
		}
		return err
	})

	return files
}

func archiveToZip(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {

		if len(call.ArgumentList) != 2 {
			logger.Panic("zip Should have two parameters. The first is the name of the zip, The seconds is the name of the folder to zip")
		}
		v := call.ArgumentList[0]
		if !v.IsString() {
			logger.Panic("Invalid zip file name")
		}
		zipFileName := v.String()

		v = call.ArgumentList[1]
		if !v.IsString() {
			logger.Panic("Invalid path to zip")
		}
		sourcePath := v.String()

		zippedFiles, err := fsutils.ZipFolder(zipFileName, sourcePath)
		if err != nil {
			logger.Panic("Fialed to zip folder: %s, %s", sourcePath, err)
		}
		logger.Infof("Zipped files: %v", zippedFiles)
		v, _ = vm.ToValue(zippedFiles)
		return v
	}
}

func unarchiveFromZip(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) != 2 {
			logger.Panic("unzip Should have two parameters. The first is the name of the zip to extract, The seconds is the destination folder.")
		}
		v := call.ArgumentList[0]
		if !v.IsString() {
			logger.Panic("Invalid zip file name")
		}
		zipFileName := v.String()

		v = call.ArgumentList[1]
		if !v.IsString() {
			logger.Panic("Invalid path to zip")
		}
		targetPath := v.String()

		fileList, err := fsutils.Unzip(zipFileName, targetPath)
		if err != nil {
			logger.Panic("Fialed to unzip file: ", err)
		}

		logger.Infof("Unzipped files %v", fileList)

		v, _ = vm.ToValue(fileList)
		return v
	}

}

//RegisterObjects register archive objects
func RegisterObjects(vm *otto.Otto, lgr *logrus.Logger) {
	logger = lgr
	vm.Set("arZip", archiveToZip(vm))
	vm.Set("arUnzip", unarchiveFromZip(vm))
}

func exampleImplementation(vm *otto.Otto) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		return otto.Value{}
	}
}
