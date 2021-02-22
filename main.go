package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/robertkrimen/otto"
	"github.com/sagiforbes/banai/commands/archive"

	"github.com/sagiforbes/banai/commands/fs"
	hashImpl "github.com/sagiforbes/banai/commands/hash"
	"github.com/sagiforbes/banai/commands/shell"
	"github.com/sirupsen/logrus"
)

const (
	defaultScriptFileName = "Banaifile"
	mainFuncName          = "main"
)

func loadScript(fileName string) string {
	b, e := ioutil.ReadFile(fileName)
	if e != nil {
		if os.IsNotExist(e) {
			panic(fmt.Sprint("Script file: " + fileName + ", not found"))
		} else {
			panic(e)
		}

	}
	return string(b)
}

func runBuild(scriptFileName string, funcCalls []string, outputConsummer *chan string) (done chan bool, abort chan bool, startErr error) {
	abort = make(chan bool)
	done = make(chan bool)

	var logger *logrus.Logger = logrus.New()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				if outputConsummer != nil {
					*outputConsummer <- fmt.Sprint(err)
				}
				logger.Error(err)
				logger.Error("Script execution exit with error !!!!!")
			}

			done <- true

		}()
		vm := otto.New()
		vm.Interrupt = make(chan func(), 1)
		go func() {
			<-abort
			vm.Interrupt <- func() { panic("Abort execution") }
		}()
		if scriptFileName == defaultScriptFileName {
			_, err := os.Stat(scriptFileName)
			if os.IsNotExist(err) {
				scriptFileName = defaultScriptFileName + ".js"
			}
		}
		script, err := vm.Compile(scriptFileName, loadScript(scriptFileName))
		if err != nil {
			panic(fmt.Sprintln("Failed to compile script ", err))
		}
		_, err = vm.Run(script)
		if err != nil {
			logger.Panic("Error running script", err)
		}
		shell.RegisterObjects(vm, logger)
		archive.RegisterObjects(vm, logger)
		fs.RegisterObjects(vm, logger)
		hashImpl.RegisterObjects(vm, logger)

		if len(funcCalls) > 0 {
			var funcVal otto.Value
			for _, f := range funcCalls {
				funcVal, err = vm.Get(f)
				if err != nil || funcVal == otto.UndefinedValue() {
					logger.Error("Cannot execute finction", f, err)
					panic(err)
				} else {
					logger.Println("Executing", f)
					_, err = funcVal.Call(funcVal)
					if err != nil {
						logger.Panic("Failed when executing javascript ", err)
					}
				}
			}

		} else {
			mainFunc, err := vm.Get(mainFuncName)
			if err != nil || mainFunc == otto.UndefinedValue() {
				logger.Warn("No main function defined")
			} else {
				logger.Println("Starting", mainFuncName)
				_, err := mainFunc.Call(mainFunc)
				if err != nil {
					logger.Panic("Failed when executing javascript ", err)
				}
			}

		}
	}()
	return
}

func main() {

	var scriptFileName = defaultScriptFileName
	var funcCalls []string
	var isAgent bool
	flag.StringVar(&scriptFileName, "f", defaultScriptFileName, "Set script to run. Default is Banaifile")
	flag.BoolVar(&isAgent, "agent", false, "true if banai is run as agent")
	flag.Parse()

	funcCalls = flag.Args()

	//----------- converting
	if !isAgent {
		doneCH, _, _ := runBuild(scriptFileName, funcCalls, nil)

		<-doneCH
		fmt.Println("Done running Banaifile", scriptFileName)
	}

}
