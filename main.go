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
	"github.com/sagiforbes/banai/infra"
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

	var b = infra.NewBanai()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				if outputConsummer != nil {
					*outputConsummer <- fmt.Sprint(err)
				}
				b.Logger.Error(err)
				b.Logger.Error("Script execution exit with error !!!!!")
			}
			b.Close()
			done <- true

		}()

		go func() {
			<-abort
			b.Jse.Interrupt <- func() { panic("Abort execution") }
		}()
		if scriptFileName == defaultScriptFileName {
			_, err := os.Stat(scriptFileName)
			if os.IsNotExist(err) {
				scriptFileName = defaultScriptFileName + ".js"
			}
		}
		script, err := b.Jse.Compile(scriptFileName, loadScript(scriptFileName))
		if err != nil {
			panic(fmt.Sprintln("Failed to compile script ", err))
		}
		_, err = b.Jse.Run(script)
		if err != nil {
			b.Logger.Panic("Error running script", err)
		}
		shell.RegisterJSObjects(b)
		archive.RegisterJSObjects(b)
		fs.RegisterJSObjects(b)
		hashImpl.RegisterJSObjects(b)

		if len(funcCalls) > 0 {
			var funcVal otto.Value
			for _, f := range funcCalls {
				funcVal, err = b.Jse.Get(f)
				if err != nil || funcVal == otto.UndefinedValue() {
					b.Logger.Error("Cannot execute finction", f, err)
					panic(err)
				} else {
					b.Logger.Println("Executing", f)
					_, err = funcVal.Call(funcVal)
					if err != nil {
						b.Logger.Panic("Failed when executing javascript ", err)
					}
				}
			}

		} else {
			mainFunc, err := b.Jse.Get(mainFuncName)
			if err != nil || mainFunc == otto.UndefinedValue() {
				b.Logger.Warn("No main function defined")
			} else {
				b.Logger.Println("Starting", mainFuncName)
				_, err := mainFunc.Call(mainFunc)
				if err != nil {
					b.Logger.Panic("Failed when executing javascript ", err)
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
