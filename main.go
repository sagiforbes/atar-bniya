package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dop251/goja"
	"github.com/sagiforbes/banai/commands/archive"
	"github.com/sagiforbes/banai/commands/fs"
	hashImpl "github.com/sagiforbes/banai/commands/hash"
	"github.com/sagiforbes/banai/commands/httpclient"
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
			b.Jse.Interrupt("Abort execution")
		}()
		if scriptFileName == defaultScriptFileName {
			_, err := os.Stat(scriptFileName)
			if os.IsNotExist(err) {
				scriptFileName = defaultScriptFileName + ".js"
			}
		}
		program, err := goja.Compile(scriptFileName, loadScript(scriptFileName), false)
		if err != nil {
			panic(fmt.Sprintln("Failed to compile script ", scriptFileName, err))
		}

		shell.RegisterJSObjects(b)
		archive.RegisterJSObjects(b)
		fs.RegisterJSObjects(b)
		hashImpl.RegisterJSObjects(b)
		httpclient.RegisterJSObjects(b)

		runVal, err := b.Jse.RunProgram(program)

		if err != nil {
			b.Logger.Panic("Failed to run program", err)
		}
		b.Logger.Info("Programe run return value: ", runVal)
		if len(funcCalls) > 0 {
			for _, fn := range funcCalls {
				fnc, ok := goja.AssertFunction(b.Jse.Get(fn))
				if !ok {
					b.Logger.Panic(fmt.Errorf("function %s not found", fn))
				}
				fnc(goja.Undefined())
			}

		} else {
			mainFunc, ok := goja.AssertFunction(b.Jse.Get("main"))
			if !ok {
				b.Logger.Panic("main function not found")
			}
			mainFunc(goja.Undefined())

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
