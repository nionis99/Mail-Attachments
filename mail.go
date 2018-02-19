package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/jhillyerd/enmime"
	pop3 "github.com/taknb2nch/go-pop3"
)

var (
	Debug      bool
	configFile string
	root       string
	data       string
	md         string
	dir        string
	num        int
)

func main() {
	Config()
	getAttach()

}
func Config() {
	flag.StringVar(&configFile, "c", "", "config file path")
	flag.Parse()
	CfgParse(configFile)
	root = CfgString("root")
	if CfgHasKey("debug") {
		Debug = CfgBool("debug")
	}
}
func getAttach() {
	client, err := pop3.Dial(CfgString("address"))
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	defer func() {
		client.Quit()
		client.Close()
	}()
	if err = client.User(CfgString("user")); err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
	if err = client.Pass(CfgString("pass")); err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
	if err := pop3.ReceiveMail(CfgString("address"), CfgString("user"), CfgString("pass"),
		func(number int, uid, data string, err error) (bool, error) {
			env, err := enmime.ReadEnvelope(bytes.NewBuffer([]byte(data)))
			if err != nil {
				panic(err)
			}
			num = number
			t := time.Now()
			data = t.Format("20060102-1504")
			if len(env.Attachments) >= 1 {
				for _, Att := range env.Attachments {
					b, err := ioutil.ReadAll(Att)
					if err != nil {
						log.Fatal(err)
					}
					if CfgBool("dir") {
						md = CfgString("root") + data
						os.Mkdir(md, 0777)
						dir = md + "/" + Att.FileName
						ioutil.WriteFile(dir, b, 0777)
						if err = client.Dele(number); err != nil {
							log.Printf("Error: %v\n", err)
						}

					} else {
						dir = CfgString("root") + Att.FileName
						ioutil.WriteFile(dir, b, 0777)
						if err = client.Dele(number); err != nil {
							log.Printf("Error: %v\n", err)
						}
					}
				}
			}
			return false, nil
		}); err != nil {
		log.Fatalf("%v\n", err)
	}

}
