// hijack project main.go
package main

import (
	"bufio"
	//"flag"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/pkg/term"
	//"github.com/docker/docker/utils"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	//"syscall"
	daemon "hijack"
	"time"
	//"unsafe"
)

const (
	version = "V0.1.0"
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

func connect(reurl string) (net.Conn, *bufio.Reader, error) {
	turl, _ := url.Parse(reurl)
	req, err := http.NewRequest("POST", reurl, nil)
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "tcp")

	//laddr, _ := net.ResolveTCPAddr("tcp", ":9999")
	//raddr, _ := net.ResolveTCPAddr("tcp", "180.175.173.166:80")
	//con, err := net.DialTCP("tcp", laddr, raddr)
	con, err := net.Dial("tcp", turl.Host)
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}
	clientcon := httputil.NewClientConn(con, nil)
	clientcon.Do(req)
	com, br := clientcon.Hijack()
	return com, br, nil
}

type broadcast struct {
	listerners []io.Writer
}

func newbro() *broadcast {
	return &broadcast{}
}
func (b *broadcast) appendbro(w io.Writer) {
	b.listerners = append(b.listerners, w)
}

func (b *broadcast) Write(data []byte) (int, error) {
	if len(b.listerners) <= 0 {
		return 0, fmt.Errorf("no listerners!")
	}
	for _, v := range b.listerners {
		//fmt.Println(v)
		if _, err := v.Write(data); err != nil {
			return 0, err
		}

	}
	return len(data), nil
}
func du(com *broadcast, br *bufio.Reader, master bool) {
	fd, _ := term.GetFdInfo(os.Stdin)
	//sta := &term.Termios{}
	//syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(sta)))
	//sta.Lflag ^= syscall.ICANON | syscall.ECHONL
	//syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCSETS, uintptr(unsafe.Pointer(sta)))
	oldState, err := term.SetRawTerminal(fd)
	if err != nil {
		return
	}
	defer term.RestoreTerminal(fd, oldState)

	go func() {
		_, err := io.Copy(com, os.Stdin)
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
	if master {

		_, err := io.Copy(os.Stdout, br)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	fmt.Println("du end")
}
func main() {

	app := cli.NewApp()
	app.Usage = "并发连接终端"
	app.Version = version
	app.Author = "leejan"
	app.Email = "lijian@bsgchina.com"

	//app.Flags = []cli.Flag{
	//	cli.StringFlag{
	//		Name:  "master",
	//		Usage: "set main connet",
	//		Value: "127.0.0.1:18002",
	//	},
	//	cli.StringSliceFlag{
	//		Name:  "slaves",
	//		Usage: "set slaves connect",
	//		//Value: &cli.StringSlice{"127.0.0.1:18002"},
	//	},
	//}

	//app.Before = func(c *cli.Context) error {

	//	return nil
	//}

	app.Commands = []cli.Command{
		{
			Name:      "cli",
			ShortName: "link",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "master,m", Value: "127.0.0.1", Usage: "write the configuration to the specified file"},
				cli.StringFlag{
					Name:  "slaves,s",
					Usage: "set slaves connect",
					//Value: &(cli.StringSlice{}),
				},
			},
			Usage: "start server",
			Action: func(c *cli.Context) {
				fmt.Println(c.Args())
				bro := newbro()
				if src := c.String("slaves"); src != "" {
					slaves := strings.Split(src, ",")

					if len(slaves) > 0 {
						for _, v := range slaves {
							com, _, err := connect("http://" + v + ":18003/cmd")
							if err != nil {
								continue
							}
							bro.appendbro(com)
							fmt.Println("append", v)
						}
					}
				}

				master := c.String("master")
				fmt.Println("master", master)
				com, br, err := connect("http://" + master + ":18003/cmd")
				if err != nil {
					return
				}
				bro.appendbro(com)
				du(bro, br, true)
				fmt.Println("end")
			},
		},
		{
			Name:      "daemon",
			ShortName: "d",
			Usage:     "start server",
			Action: func(c *cli.Context) {
				daemon.Main()
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		return
	}
	//one = sync.Once{}
	//fmt.Println(os.TempDir())
	//turl, _ := url.Parse(reurl)
	//fmt.Println(turl.Host)
	//req, err := http.NewRequest("POST", reurl, nil)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}

	//laddr, _ := net.ResolveTCPAddr("tcp", ":9999")
	//raddr, _ := net.ResolveTCPAddr("tcp", "180.175.173.166:80")
	//con, err := net.DialTCP("tcp", laddr, raddr)

	//con, err := net.Dial("tcp", turl.Host)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//clientcon := httputil.NewClientConn(con, nil)
	//clientcon.Do(req)
	//com, br := clientcon.Hijack()

	//cle := new(clientstdin)
	//cle.hostname = turl.Host
	//cle.chin = make(chan int, 1)
	//cle.chin <- 1
	//cle.chout = make(chan int, 1)
	//fd, _ := term.GetFdInfo(os.Stdout)
	//stat, _ := term.SaveState(fd)
	//term.DisableEcho(fd, stat)

	//go func() {
	//	//com.Write([]byte("\n"))
	//	_, err = io.Copy(com, os.Stdin)
	//	if err != nil {
	//		fmt.Println(err)
	//		return
	//	}
	//	//io.Copy(com, os.Stdin)
	//}()
	////stdcopy.StdCopy(os.Stdout, os.Stderr, br)
	//io.Copy(os.Stdout, br)
	//fmt.Println("Hello World!")
}
