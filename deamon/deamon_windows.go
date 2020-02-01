package deamon

import(
	"os"
	"fmt"
)

type Deamon struct {
}
type DeamonParam struct {
	Sta bool
}

func init(){
	fmt.Println("The deamon not support windows")
	os.Exit(1)
}

func (d *Deamon)StartFinish(dp *DeamonParam,sta *bool)(err error){
	*sta = true
	return nil
}

func (d *Deamon)StopFinish(dp *DeamonParam,sta *bool)(err error){
	*sta = true
	return nil
}

func DeamonHandleSignals(startHandler,stopHandler func()bool) {
	return
}