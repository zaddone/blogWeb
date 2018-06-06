package main
import (
	"fmt"
	"log"
	"os"
	"flag"
	"os/exec"
	"path/filepath"
	"bytes"
	"time"
	"strings"
	"./emailServer"
)
type runFunc func()error
var (
	Site  = flag.String("site","https://zaddone.github.io","site addr")
	Source= flag.String("source","/home/dimon/Documents/blog_web","filesystem path to read files relative from")
	Theme  = flag.String("theme","hugo-bootstrap","theme name")
	runMap map[string]runFunc
)
const (
	public string = "public"
)
func StartCmd(cmd * exec.Cmd) (err error) {
	var _out bytes.Buffer
	var _err bytes.Buffer
	cmd.Stderr = &_err
	cmd.Stdout = &_out
	if err = cmd.Start();err != nil {
		log.Println(cmd.Args)
		return err
	}
	cmd.Wait()
	if _err.Len() >0 {
		err = fmt.Errorf(_err.String())
	}
	log.Println(cmd.Args)
	//if _out.Len()>0 {
	//	log.Println(_out.String())
	//}
	return err

}
func deploy() error {
	return StartCmd(exec.Command("hugo",
			fmt.Sprintf("--theme=%s",*Theme),
			fmt.Sprintf("--baseUrl=%s",*Site),
			fmt.Sprintf("--source=%s",*Source)))
}
func Push() (err error) {
	err = os.Chdir(filepath.Join(*Source,public))
	if err != nil {
		return err
	}
	err = StartCmd(exec.Command("git","add","."))
	if err != nil {
		return err
	}
	err = StartCmd(exec.Command("git","commit","-m","automatic"))
	if err != nil {
		return err
	}
	err = StartCmd(exec.Command("git","push","origin","master"))
	if err != nil {
		log.Println(err)
	}
	return nil
}
func Update() (err error) {
	err = deploy()
	if err != nil {
		return err
	}
	return Push()
}

func Add() (err error) {

	err = os.Chdir(*Source)
	if err != nil {
		return err
	}
	now := time.Now()
	return StartCmd(exec.Command("hugo","new",
	filepath.Join("post",fmt.Sprintf("%d",now.Year()),
	fmt.Sprintf("%s_%d.md",strings.ToLower(now.Month().String()),now.Day()))))

}

func init(){
	flag.Parse()
	runMap = make(map[string]runFunc)
	runMap["update"] = Update
	runMap["add"] = Add

	emailServer.Monitor(Update)
}

func main(){

	var err error
	var n int
	var cmd string
	for{
		n,err = fmt.Scanf("%s",&cmd)
		if n ==0 {
			continue
		}
		f := runMap[cmd]
		if f == nil {
			log.Println(cmd,n,err)
		}else{
			err = f()
			if err != nil {
				log.Println(err)
			}
		}
	}

}
