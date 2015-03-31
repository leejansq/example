// hijack project main.go
package main

import (
	"fmt"
	//"github.com/docker/docker/utils"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

func hijack(w http.ResponseWriter, r *http.Request) {
	fmt.Println("/cmd")
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = io.Copy(conn, os.Stdin)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("end")
}

var reurl string = "http://127.0.0.1:8002/cmd"

type clientstdin struct {
	sync.Mutex
	flag     bool
	wn       int
	hostname string
	chin     chan int
	chout    chan int
}

func (cl *clientstdin) Read(b []byte) (int, error) {
	//cl.Lock()
	<-cl.chin
	defer func() {
		cl.chout <- 1
	}()

	//fmt.Println("stdin.. ", cl.flag)
	//defer cl.Unlock()
	if cl.flag == false {
		n, err := os.Stdin.Read(b)
		fmt.Println("<<<<", string(b), "<<<<")
		return n, err
	} else {
		//cl.flag = false
		time.Sleep(time.Second)
		b = []byte("pwd\n")
		fmt.Println("<<<<", string(b), "<<<<")
		return 4, nil
	}

}

func (cl *clientstdin) Write(b []byte) (int, error) {
	<-cl.chout
	one.Do(func() {
		fmt.Println("once")
		cl.chout <- 1
		return
	})
	defer func() {
		cl.chin <- 1
	}()
	//fmt.Println("stdout.. ", cl.flag)
	if cl.flag == false {
		cl.flag = true
		n, err := os.Stdout.Write(b)
		//fmt.Println(">>>", string(b[:n]),">>>>")
		return n, err
	} else {
		cl.flag = false
		sy := "\n[" + cl.hostname + "|" + strings.TrimRight(string(b), "\n") + "]$ "
		os.Stdout.Write([]byte(sy))
		//fmt.Println("$$$$$", sy)
		//b=[]byte("pwd\n")
		//cl.chin <- 1
		return len(b), nil
	}

}

var one sync.Once

func main() {
	one = sync.Once{}
	fmt.Println(os.TempDir())
	turl, _ := url.Parse(reurl)
	fmt.Println(turl.Host)
	req, err := http.NewRequest("POST", reurl, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	con, err := net.Dial("tcp", turl.Host)
	if err != nil {
		fmt.Println(err)
		return
	}
	clientcon := httputil.NewClientConn(con, nil)
	clientcon.Do(req)
	com, br := clientcon.Hijack()
	//cle := new(clientstdin)
	//cle.hostname = turl.Host
	//cle.chin = make(chan int, 1)
	//cle.chin <- 1
	//cle.chout = make(chan int, 1)

	go func() {
		//com.Write([]byte("\n"))
		_, err = io.Copy(com, os.Stdin)
		if err != nil {
			fmt.Println(err)
			return
		}
		//io.Copy(com, os.Stdin)
	}()
	io.Copy(os.Stdout, br)
	fmt.Println("Hello World!")
}
