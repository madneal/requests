package requests

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type Request struct {
	Url     string
	Headers map[string]string
	Method  string
}

func init() {
	f, err := os.Open("config.yaml")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		fmt.Println(err)
	}

}

func main() {

}
