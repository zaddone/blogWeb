package emailServer
import(

	"log"
	//"bytes"
	"github.com/BurntSushi/toml"
	"os"

)

const (
	FileName string = "config.info"
)
type Config struct {

	Imap string
	User string
	Pass string
	MBox string
	BlogBox string

	MsgCount int
	Enduid int

}

func (self *Config) Save() {

	//var buf bytes.Buffer
	fi,err := os.OpenFile(FileName,os.O_CREATE|os.O_WRONLY,0777)
	//fi,err := os.Open(FileName)
	if err != nil {
		log.Fatal(err)
	}
	defer fi.Close()
	e := toml.NewEncoder(fi)
	err = e.Encode(self)
	if err != nil {
		log.Fatal(err)
	}

}
func NewConfig()  *Config {

	var c Config
	_,err := os.Stat(FileName)
	if err != nil {
		c.BlogBox = "blog"
		c.MBox = "INBOX"
		c.Imap = "imap.qq.com:993"
		c.User = "zaddone@qq.com"
		c.Pass = "omngxacjppjucaae"
		c.MsgCount = 1
	}else{
		if _,err := toml.DecodeFile(FileName,&c);err != nil {
			log.Fatal(err)
		}
	}
	return &c

}
