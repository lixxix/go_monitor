package main

import (
	"fmt"
	"log"
	"monitor/proc"
	"monitor/utils"
	"monitor/watcher"
	"os"
	"strconv"
)

// func signalListen() {
// 	c := make(chan os.Signal)
// 	signal.Notify(c, syscall.SIGKILL)
// 	for {
// 		s := <-c
// 		//收到信号后的处理，这里只是输出信号内容，可以做一些更有意思的事
// 		fmt.Println("get signal:", s)
// 	}
// }

func main() {
	directroy, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("directroy:", directroy)

	Check := make(chan struct{})

	ww := watcher.CreateWatcher()
	ww.Init()
	fmt.Println(ww.CheckExitsPid)
	starter := proc.CreateProcStarter(os.Environ())
	go starter.Run()
	// 将进程数据给到Proc
	starter.CheckExistsProcess(ww.CheckExitsPid)
	for k, v := range ww.MapExec {
		starter.PushProcess(utils.ProcessData{
			Port: k,
			Exec: v,
			Args: []string{"-port", strconv.Itoa(k)},
		})
	}

	go ww.Watching(Check, starter.StartChan)
	starter.MonitorRes(Check)

	// time.Sleep(time.Minute * 10)
	// ww.CheckNetListen()
	// env := os.Environ()
	// out_f, err := os.Create(fmt.Sprintf("%d_%s.log", 37711, time.Now().Format("20060102150405")))
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
	// err_f, err := os.Create(fmt.Sprintf("%d_%s-error.log", 37711, time.Now().Format("20060102150405")))
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
	// procAttr := &os.ProcAttr{
	// 	Dir: "E:\\go\\src\\kirin_go_combine\\game\\pochiko\\",
	// 	Env: env,
	// 	Sys: &syscall.SysProcAttr{
	// 		HideWindow: false,
	// 	},
	// 	Files: []*os.File{
	// 		nil,
	// 		out_f,
	// 		err_f,
	// 	},
	// }

	// pid, err := os.StartProcess("pochiko.exe", []string{"pochiko.exe", "-port", "37711"}, procAttr)
	// if err != nil {
	// 	fmt.Printf("Error %v starting process!", err) //
	// 	os.Exit(1)
	// }
	// fmt.Printf("The process id is %v\n", pid)
	// go signalListen()

	// time.Sleep(time.Second * 20)
	// pid.Kill()
}
