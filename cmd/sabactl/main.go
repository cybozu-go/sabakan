package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
)

var (
	flagServer = flag.String("server", "http://localhost:8888", "<Listen IP>:<Port number>")
)

func main() {

	flag.Parse()
	req, err := http.NewRequest("GET", *flagServer+"/hello", nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	req = req.WithContext(context.Background())

	client := &cmd.HTTPClient{
		Client:   &http.Client{},
		Severity: log.LvDebug,
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	buf := new(bytes.Buffer)
	io.Copy(buf, res.Body)
	ret := buf.Bytes()
	fmt.Println(string(ret))
}
