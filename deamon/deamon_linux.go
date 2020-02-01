package deamon

import(
	"os"
	"fmt"
	"net"
	"time"
	"flag"
	"strconv"
	"syscall"
	"os/exec"
	"net/rpc"
	"net/http"
	"io/ioutil"
	"os/signal"
)
type Deamon struct {
}
type DeamonParam struct {
	Sta bool
}

var(
	addr = "127.0.0.255:20111"
	runFlag = make(chan bool)
	startChan = make(chan bool)
	stopChan = make(chan bool)
	runParam = flag.String("s","start", 
	`
	run app as a daemon with -s start
	stop app with -s stop`)
)
func (d *Deamon)run(){
	go func(){
		rpc.Register(new(Deamon))
		rpc.HandleHTTP()
		listen, err := net.Listen("tcp",addr)
		if err != nil {
			fmt.Println(err)
			runFlag <- false
			return
		}
		runFlag <- true
		http.Serve(listen, nil)
	}()
}

func (d *Deamon)StartFinish(dp *DeamonParam,sta *bool)(err error){
	startChan <- dp.Sta
	*sta = true
	return nil
}

func (d *Deamon)StopFinish(dp *DeamonParam,sta *bool)(err error){
	stopChan <- dp.Sta
	*sta = true
	return nil
}

func startFinish(sta bool)bool{
	var result bool
	conn, err := rpc.DialHTTP("tcp",addr)
	if err != nil {
		fmt.Println(err)
		return false
	}
	err = conn.Call("Deamon.StartFinish",&DeamonParam{Sta:sta}, &result)
	if nil != err{
		return false
	}
	conn.Close()
	return true
}

func stopFinish(sta bool)bool{
	var result bool
	conn, err := rpc.DialHTTP("tcp",addr)
	if err != nil {
		fmt.Println(err)
		return false
	}
	err = conn.Call("Deamon.StopFinish",&DeamonParam{Sta:sta}, &result)
	if nil != err{
		return false
	}
	conn.Close()
	return true
}

func DeamonHandleSignals(startHandler,stopHandler func()bool) {
	var sig os.Signal
	signalChan := make(chan os.Signal)
	signal.Notify(
		signalChan,
		syscall.SIGUSR1,
	)
	sta := startHandler()
	startFinish(sta)
	if false == sta{
		os.Exit(0)
	}
	for {
		sig = <-signalChan
		if syscall.SIGUSR1 == sig {
			sta = stopHandler()
			stopFinish(sta)
			os.Exit(0)
		}
	}
}

func init() {
	if !flag.Parsed() {
		flag.Parse()
	} 
	deamon := &Deamon{}
	switch *runParam{
		case "start":{
			deamon.run()
			sta := <- runFlag
			close(runFlag)
			if true == sta{
				cmd := exec.Command(os.Args[0]) 
				go func(){
					ticker := time.NewTicker(10*time.Second)//wait most 10S
					for{
						select{
							case sta,_:=<-startChan:{
								if true == sta{
									ioutil.WriteFile(os.Args[0]+".pid", []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0666)
									fmt.Printf("%s [PID] %d running!\n", os.Args[0],cmd.Process.Pid)
								}else{
									fmt.Printf("%s [PID] %d start fail!\n", os.Args[0],cmd.Process.Pid)
								}
								os.Exit(0)  
							}
							case <-ticker.C:{
								ticker.Stop()
								fmt.Printf("%s [PID] %d start over time!\n", os.Args[0],cmd.Process.Pid)
								os.Exit(0) 
							}
						}
					}
				}()
				fmt.Printf("%s [PID] start...\n", os.Args[0])
				cmd.Run()
			}
			return
		}
		case "stop":{
			strb, err := ioutil.ReadFile(os.Args[0]+".pid")
			if nil != err{
				fmt.Println(err)
				os.Exit(0)
			}
			pid,err := strconv.Atoi(string(strb))
			if nil != err{
				os.Exit(0)
			}
			pro, err := os.FindProcess(pid)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
			deamon.run()
			sta := <- runFlag
			close(runFlag)
			if true == sta{
				fmt.Printf("%s [PID] %d stoping...\n", os.Args[0],pid) 
				err := pro.Signal(syscall.SIGUSR1)
				if nil != err{
					fmt.Println(err)
					os.Exit(0)
				}
				for{
					ticker := time.NewTicker(10*time.Second)//wait most 10S
					for{
						select{
							case sta,_ := <-stopChan:{
								if true == sta{
									fmt.Printf("%s [PID] %d stop success!\n", os.Args[0],pid)
								}else{
									fmt.Printf("%s [PID] %d stop fail!\n", os.Args[0],pid)
								}
								os.Exit(0)  
							}
							case <-ticker.C:{
								ticker.Stop()
								fmt.Printf("%s [PID] %d stop over time!\n", os.Args[0],pid)
								os.Exit(0) 
							}
						}
					}
				}
			}
			return
		}
	}
	fmt.Println(
	`
	run app as a daemon with -s start
	stop app with -s stop`)
	os.Exit(0)
}