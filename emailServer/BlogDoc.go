package emailServer

import(
	"os"
	"path/filepath"
	"strings"
	"time"
	"log"
	"fmt"
	//"io"
)

type BlogDoc struct {
	Title string
	Date time.Time
	Body string
	slug string
	//slug,title,date
	//---
	//date: 2018-05-15T12:56:46+08:00
}
func (self *BlogDoc) content (w *os.File){
	br :="\n"
	_,err := w.WriteString("---"+br)
	if err != nil {
		log.Fatal(err)
	}
	_,err = w.WriteString("slug: "+self.slug+br)
	if err != nil {
		log.Fatal(err)
	}
	_,err = w.WriteString("title: "+self.Title+br)
	if err != nil {
		log.Fatal(err)
	}
	_,err = w.WriteString("date: "+self.Date.Format("2006-01-02T15:04:05+08:00")+br)
	if err != nil {
		log.Fatal(err)
	}


	_,err = w.WriteString("---"+br)
	if err != nil {
		log.Fatal(err)
	}
	_,err = w.WriteString(self.Body)
	if err != nil {
		log.Fatal(err)
	}
}
func (self *BlogDoc) Save(){

	fileName:=fmt.Sprintf("%s_%d",strings.ToLower(self.Date.Month().String()),self.Date.Day())
	self.slug = fileName

	filePath := filepath.Join("content","post",fmt.Sprintf("%d",self.Date.Year()))
	_,err := os.Stat(filePath)
	if err != nil {
		err = os.MkdirAll(filePath,0777)
		if err != nil {
			log.Fatal(err)
		}
	}
	fn := filepath.Join(filePath,self.slug+".md")
	n:=0
	var f *os.File
	for{
		_,err := os.Stat(fn)
		if err != nil {
			f,err = os.OpenFile(fn,os.O_CREATE|os.O_APPEND|os.O_WRONLY,0777)
			//self.slug =fileName
			if err != nil {
				log.Fatal(err)
			}
			self.content(f)
			f.Close()
			break
		}
		n++
		self.slug = fmt.Sprintf("%s_%d",fileName,n)
		fn = filepath.Join(filePath,self.slug+".md")
	}




}
