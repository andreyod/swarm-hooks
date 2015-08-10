package authentication

import (
	"encoding/json"
	
	"log"
	"os"
)

type Configs struct {
	AuthTokenHeader string
	TenancyLabel       string
	KeystoneUrl        string
	KeyStoneXAuthToken string
}

const defaultConfigurationFileCreationPath = "/tmp/authHookConf.json"

var Configuration *Configs

func (*Configs) ReadConfigurationFormfile() {
	file, _ := os.Open("authHookConf.json")
	decoder := json.NewDecoder(file)
	Configuration = new(Configs)
	err := decoder.Decode(&Configuration)
	if err != nil {
		log.Println("error:", err)
	}
	log.Println("*************************")
	log.Println(Configuration)
	log.Println("*************************")
}

func (*Configs) GetConf() *Configs {
	return Configuration
}

func (*Configs) CreateDefaultsConfigurationfile() {
	confs := Configs{
		AuthTokenHeader: "X-Auth-Token",
		TenancyLabel:       "com.ibm.tenant.0",
		KeystoneUrl:        "http://127.0.0.1:5000/v2.0/",
		KeyStoneXAuthToken: "ADMIN",
	}

	bytesConfigurationData, e0 := json.Marshal(&confs)

	if e0 != nil {
		log.Fatal(e0)
	}
	f, e1 := os.Create(defaultConfigurationFileCreationPath)
	if e1 != nil {
		log.Fatal(e1)
	}

	defer f.Close()
	n2, e2 := f.Write(bytesConfigurationData)
	if e2 != nil {
		log.Fatal(e2)
	}
	log.Println(n2)
	f.Sync()

}
