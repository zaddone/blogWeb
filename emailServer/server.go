package emailServer

import(
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
//	"fmt"
	"strings"
	//"flag"
	"log"
	//"net/mail"
	"io/ioutil"
	"io"
	"time"
	//"sync"
)
const (
	subjectTag = "blog "
	contentType = "text/plain"
	MonitorSpaceSecond= 1

)

type EmailClient struct {

	c    *client.Client
	config *Config
	Update bool

}

func NewEmailClient() (*EmailClient) {

	conf := NewConfig()
	c,err := client.DialTLS(conf.Imap,nil)
	if err != nil {
		log.Fatal(err)
	}
	if err := c.Login(conf.User,conf.Pass);err != nil {
		log.Fatal(err)
	}

	return &EmailClient{
		c:c,
		config:conf}

}
func (self *EmailClient) Close(){
	err := self.c.Logout()
	if err != nil {
		log.Fatal(err)
	}
}

func (self *EmailClient) GetSeqSet(box *imap.MailboxStatus) *imap.SeqSet{

	if (int(box.Messages) - self.config.MsgCount) < 1 {
		return nil
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(uint32(self.config.MsgCount+1),box.Messages)
	self.config.MsgCount = int(box.Messages)
	self.config.Save()
	return seqset

}
func (self *EmailClient) readBody(seqset *imap.SeqSet,docs []*BlogDoc){

	section := &imap.BodySectionName{}
	//items := []imap.FetchItem{section.FetchItem()}
	items := []imap.FetchItem{section.FetchItem(),imap.FetchEnvelope}

	messages := make(chan *imap.Message, len(docs))
	done := make(chan error, 1)
	go func() {
		done <- self.c.Fetch(seqset, items, messages)
	}()
	j:=0
	for msg := range messages {
	//msg := <-messages
		//msg.Format()
		subject,err := charset.DecodeHeader(msg.Envelope.Subject)
		doc := docs[j]
		if doc == nil {
			log.Fatal("docs d is nil")
		}
		log.Println(subject,doc.Title,err)
		//for k,v := range msg.Body {
		//	log.Println(k,v)
		//}
		r := msg.GetBody(section)
		if r == nil {
			log.Println("Server didn't returned message body")
			continue
		}

		m,err := mail.CreateReader(r)
		if err != nil {
			log.Println(err)
			continue
		}
		defer m.Close()

		for {
			p,err := m.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}
			//fmt.Println(p.Header[])
			switch h := p.Header.(type) {
			case mail.TextHeader:
				Type,_,err := h.ContentType()
				if err == nil && strings.HasPrefix(Type,contentType) {
					b, _ := ioutil.ReadAll(p.Body)
					//log.Println("Got text:", string(b))
					doc.Body += string(b)
				}
			//case mail.AttachmentHeader:
			//	filename, _ := h.Filename()
			//	log.Println("Got attachment:", filename)
			}
		}
		j++
		doc.Save()
		log.Println("doc:", doc)
	}
	if err := <-done; err != nil {
		log.Println(err)
		return
	}

}

func (self *EmailClient) readSubject(seqset *imap.SeqSet){
	if seqset == nil {
		return
	}
	messages := make(chan *imap.Message,10)
	done := make(chan error,1)
	go func(){
		done <- self.c.Fetch(seqset,[]imap.FetchItem{imap.FetchEnvelope},messages)
	}()
	//var sy sync.WaitGroup
	//var bloglist []uint32
	seqsetbody := new(imap.SeqSet)
	var blogDoc []*BlogDoc
	for msg := range messages {
		subject,err := charset.DecodeHeader(msg.Envelope.Subject)
		if err != nil {
			log.Panicln(err)
			continue
		}
		if strings.HasPrefix(subject,subjectTag) {
			log.Println("* ",subject,msg.SeqNum,msg.Uid)
			blogDoc = append(blogDoc,&BlogDoc{
				Title:strings.Join(strings.Split(subject," ")[1:]," "),
				Date:msg.Envelope.Date})
			seqsetbody.AddNum(msg.SeqNum)
			//sy.Add(1)
			//go func(seqNum uint32,w *sync.WaitGroup){
			//	self.readBody(seqNum)
			//	w.Done()
			//}(msg.SeqNum,&sy)
		}else{
		log.Println("| ",subject,msg.SeqNum,msg.Uid)
		}
	}

	if err := <-done;err != nil {
		log.Fatal(err)
	}
	if len(blogDoc) >0 {
		self.readBody(seqsetbody,blogDoc)
		self.Update = true
	}
	//sy.Wait()

}

func (self *EmailClient) Read() {

	mbox,err := self.c.Select(self.config.MBox,false)
	if err != nil {
		log.Fatal(err)
	}
	self.Update = false
	self.readSubject(self.GetSeqSet(mbox))

}
func Monitor(up func() error){
	var c *EmailClient
	t := time.Tick(MonitorSpaceSecond * time.Minute)
	for{
		//log.Println(1)
		c = NewEmailClient()
		c.Read()
		if c.Update && up != nil {
			err := up()
			if err != nil {
				log.Fatal(err)
			}
		}
		c.Close()
		<-t
	}


}
//func init(){
//	go Monitor(nil)
//}
//func init(){
//
//	log.Println("Connection to server...")
//	c,err := client.DialTLS(fmt.Sprintf("imap.%s:993",*Server),nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer func(){
//		err := c.Logout()
//		if err != nil {
//			log.Fatal(err)
//		}
//	}()
//
//	log.Println("login")
//	if err := c.Login(fmt.Sprintf("%s@%s",*User,*Server),*Pass);err != nil {
//		log.Fatal(err)
//	}
//	mailboxes := make(chan *imap.MailboxInfo,10)
//	done := make(chan error,1)
//	go func(){
//		done <- c.List("","*",mailboxes)
//	}()
//	log.Println("mailboxes")
//	for m := range mailboxes {
//		log.Println("* "+m.Name)
//	}
//	if err := <-done;err != nil {
//		log.Fatal(err)
//	}
//
//	mbox,err := c.Select(MBox,false)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Println("Flags for INBOX:", mbox.Flags)
//	from := uint32(1)
//	to := mbox.Messages
//	if mbox.Messages > 3 {
//		from = mbox.Messages - 3
//	}
//	seqset := new(imap.SeqSet)
//	seqset.AddRange(from,to)
//	messages := make(chan *imap.Message,10)
//	done = make(chan error,1)
//	go func(){
//		done <- c.Fetch(seqset,[]imap.FetchItem{imap.FetchEnvelope},messages)
//	}()
//	log.Println("Last 4 messages:")
//	for msg := range messages {
//		text,err := charset.DecodeHeader(msg.Envelope.Subject)
//		log.Println("* ",text,msg.SeqNum,msg.Uid,err)
//		FetchOne(msg.SeqNum,c)
//	}
//
//	if err := <-done;err != nil {
//		log.Fatal(err)
//	}
//
//}
//func FetchOne(from uint32,c *client.Client){
//
//	seqset := new(imap.SeqSet)
//	seqset.AddRange(from, from)
//
//	// Get the whole message body
//	section := &imap.BodySectionName{}
//	items := []imap.FetchItem{section.FetchItem()}
//
//	messages := make(chan *imap.Message, 1)
//	done := make(chan error, 1)
//	go func() {
//		done <- c.Fetch(seqset, items, messages)
//	}()
//	msg := <-messages
//	r := msg.GetBody(section)
//	if r == nil {
//		log.Fatal("Server didn't returned message body")
//	}
//	if err := <-done; err != nil {
//		log.Fatal(err)
//	}
//
//	m,err := mail.CreateReader(r)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer m.Close()
//
//	for {
//
//		p,err := m.NextPart()
//		if err == io.EOF {
//			break
//		} else if err != nil {
//			log.Fatal(err)
//		}
//		//fmt.Println(p.Header[])
//		switch h := p.Header.(type) {
//		case mail.TextHeader:
//			Type,_,err := h.ContentType()
//			if err == nil && strings.Contains(Type,"text/plain") {
//				b, _ := ioutil.ReadAll(p.Body)
//				log.Println("Got text:", string(b))
//			}
//		case mail.AttachmentHeader:
//			filename, _ := h.Filename()
//			log.Println("Got attachment:", filename)
//		}
//
//	}
//	//body,err := ioutil.ReadAll(m.Body)
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//	//text,err := charset.DecodeHeader(string(body))
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//
//	//log.Println(text)
//
//}
