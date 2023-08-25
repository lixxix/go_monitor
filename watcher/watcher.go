package watcher

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"monitor/utils"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type ExecList struct {
	Exec string `yaml:"exec"`
	Port []int  `yaml:"port"`
}

func (e *ExecList) HasPort(port int) bool {
	for _, v := range e.Port {
		if v == port {
			return true
		}
	}
	return false
}

type MonitorData struct {
	Dir  string     `yaml:"dir"`
	List []ExecList `yaml:"list"`
}

func (m *MonitorData) hasPortProcess(port int) (bool, string) {
	for _, ll := range m.List {
		if ll.HasPort(port) {
			return true, ll.Exec
		}
	}
	return false, ""
}

type Watcher struct {
	netArgs       []string //网络执行命令
	monitorInfo   *MonitorData
	MapExec       map[int]string //存储端口和运行指令
	CheckExitsPid map[int]int    //存储对应的PID
	is_linux      bool
}

func (w *Watcher) CheckNetListen() {
	cmd := exec.Command("netstat", w.netArgs...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err.Error())
	}

	err = cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
	}
	s := bufio.NewScanner(out)
	for s.Scan() {
		head_str := string(s.Bytes())
		head_str = utils.MoveMoreSpace(head_str)

		splits := strings.Split(head_str, " ")

		if splits[0] == "TCP" || splits[0] == "tcp" {
			if w.is_linux {
				lastIndex := strings.LastIndex(splits[3], ":")
				if lastIndex != -1 {
					port, err := strconv.Atoi(splits[3][lastIndex+1:])
					if err != nil {
						log.Println(err.Error())
					}
					if _, ok := w.MapExec[port]; ok {
						pid_str := strings.Split(splits[len(splits)-1], "/")[0]
						pid, err := strconv.Atoi(pid_str)
						if err != nil {
							log.Println(err.Error())
						}
						w.CheckExitsPid[port] = pid
						delete(w.MapExec, port)
					}
				}
			} else {
				ip_splits := strings.Split(splits[1], ":")
				if len(ip_splits) == 2 {
					port, err := strconv.Atoi(ip_splits[1])
					if err != nil {
						log.Fatal(err.Error())
					}
					if _, ok := w.MapExec[port]; ok {
						pid, err := strconv.Atoi(splits[len(splits)-1])
						if err != nil {
							log.Println(err.Error())
						}
						w.CheckExitsPid[port] = pid
						delete(w.MapExec, port)
					}
				}
			}
		}
	}
	cmd.Wait()

	// 传递给proc进行处理
	fmt.Println(w.MapExec)
}

func (w *Watcher) Watching(check chan struct{}, procChan chan utils.ProcessData) {
	for {
		<-check

		for _, ll := range w.monitorInfo.List {
			for _, port := range ll.Port {
				w.MapExec[port] = ll.Exec
			}
		}

		log.Println("收到了检测请求")
		w.CheckNetListen()

		for k, v := range w.MapExec {
			procChan <- utils.ProcessData{
				Port: k,
				Exec: v,
				Args: []string{"-port", strconv.Itoa(k)},
			}
		}
	}
}

func (w *Watcher) Init() {
	stream, err := ioutil.ReadFile("config/monitor.yaml")
	if err != nil {
		panic(err)
	}
	w.monitorInfo = &MonitorData{}
	err = yaml.Unmarshal(stream, w.monitorInfo)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(w.monitorInfo)
		for _, ll := range w.monitorInfo.List {
			for _, port := range ll.Port {
				w.MapExec[port] = ll.Exec
			}
		}
	}

	env := os.Getenv("OS")
	if strings.HasPrefix(env, "Windows") {
		fmt.Println("Windows environment")
		w.netArgs = []string{"-aon"}
		w.is_linux = false
	} else {
		fmt.Println("unix environment netstat -tunlp")
		w.netArgs = []string{"-tunlp"}
		w.is_linux = true
	}
	w.CheckNetListen()
}

func CreateWatcher() *Watcher {
	return &Watcher{
		MapExec:       make(map[int]string),
		CheckExitsPid: make(map[int]int),
	}
}
