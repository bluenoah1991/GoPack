package main

import gopack "github.com/codemeow5/GoPack/lib"
import "time"
import "fmt"

func callback(payload []byte, err error) {
	fmt.Println(err)
}

func main() {
	opts := gopack.Options{
		Address:  "192.168.102.73:8081",
		Callback: callback,
	}
	var maxtime = time.Date(2500, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	fmt.Printf("%s", maxtime)
	gopk, err := gopack.NewGoPack(&opts)
	if err != nil {
		return
	}

	gopk.Start()

	for {
		time.Sleep(10 * time.Second)
	}
}
