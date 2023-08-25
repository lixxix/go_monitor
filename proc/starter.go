package proc

import (
	"fmt"
	"log"
	"monitor/utils"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type ProcStarter struct {
	env         []string
	StartChan   chan utils.ProcessData
	CloseChan   chan struct{}
	MapPortProc map[int]*process.Process
}

func (p *ProcStarter) startProcess(info utils.ProcessData) {
	// 创建日志文件
	_, err := os.Stat("log")
	if err != nil {
		fmt.Println(err.Error())
		os.Mkdir("log", os.ModePerm)
	} else {
		fmt.Println("exists")
	}

	if _, ok := p.MapPortProc[info.Port]; !ok {
		name := strings.Split(info.Exec, ".")[0]
		idx := strings.LastIndex(info.Exec, "/")
		if idx != -1 {
			name = strings.Split(info.Exec[idx+1:], ".")[0]
		}
		out_f, err := os.OpenFile(fmt.Sprintf("log/%s_%d_%s-out.log", name, info.Port, time.Now().Format("20060102150405")), os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err.Error())
		}
		err_f, err := os.OpenFile(fmt.Sprintf("log/%s_%d_%s-error.log", name, info.Port, time.Now().Format("20060102150405")), os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err.Error())
		}

		// cur_dir, err := os.Getwd()
		// if err != nil {
		// 	log.Println(err.Error())
		// }
		procAttr := &os.ProcAttr{
			Dir: "E:\\go\\src\\kirin_go_combine\\game\\pochiko",
			Env: p.env,
			Sys: &syscall.SysProcAttr{
				HideWindow: false,
			},
			Files: []*os.File{
				nil,
				out_f,
				err_f,
			},
		}

		proc, err := os.StartProcess(info.Exec, []string{info.Exec, "-port", strconv.Itoa(info.Port)}, procAttr)
		if err != nil {
			fmt.Printf("Error %v starting process!", err) //
		}

		// proc.Wait()

		checkMap := make(map[int]int)
		checkMap[info.Port] = int(proc.Pid)
		p.CheckExistsProcess(checkMap)
	}
}

// 将已开启的进程加入检测
func (p *ProcStarter) CheckExistsProcess(port_map map[int]int) {
	pid_map := make(map[int32]int)
	for k, v := range port_map {
		pid_map[int32(v)] = k
	}

	processes, err := process.Processes()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for _, proc := range processes {
		if _, ok := pid_map[proc.Pid]; ok {
			p.MapPortProc[pid_map[proc.Pid]] = proc
		}
	}

}

func (p *ProcStarter) PushProcess(info utils.ProcessData) {
	p.StartChan <- info
}

// 进行启动
func (p *ProcStarter) Run() {
	for {
		select {
		case info := <-p.StartChan:
			log.Println("收到启动程序")
			if _, ok := p.MapPortProc[info.Port]; !ok {
				p.startProcess(info)
			}
		case <-p.CloseChan:
			return
		}
	}
}

// 获得数据
func (p *ProcStarter) MonitorRes(check chan struct{}) {
	ticker := time.NewTicker(time.Second * 30)
	for {
		<-ticker.C
		rm := make([]int, 0)
		for port, v := range p.MapPortProc {
			merinfo, err := v.MemoryInfo()
			if err != nil {
				fmt.Println(err.Error())
				rm = append(rm, port)
				continue
			}

			if merinfo.RSS < 100_0000 {
				rm = append(rm, port)
				continue
			}

			cpu, err := v.CPUPercent()
			if err != nil {
				fmt.Println(err.Error())
				rm = append(rm, port)
				continue
			}

			memer, err := v.MemoryPercent()
			if err != nil {
				fmt.Println(err.Error())
				rm = append(rm, port)
				continue
			}

			fmt.Println(port, cpu, memer, merinfo)
		}
		if len(rm) > 0 {
			for _, port := range rm {
				delete(p.MapPortProc, port)
			}
			check <- struct{}{}
		}
	}
}

func CreateProcStarter(env []string) *ProcStarter {
	return &ProcStarter{
		env:         env,
		MapPortProc: make(map[int]*process.Process),
		StartChan:   make(chan utils.ProcessData, 10),
		CloseChan:   make(chan struct{}),
	}
}
