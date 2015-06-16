package main

import (
	"fmt"
	"os"
	//"os/exec"
	//"github.com/tfogal/ptrace"
	"encoding/hex"
	"runtime"
	"syscall"
	"time"
)

func te() {
	b := hex.EncodeToString([]byte("lee"))
	fmt.Println(string(b))
	return
}

func main() {
	//cmd := exec.Command("./child")
	//cmd.SysProcAttr = &syscall.SysProcAttr{
	//	Ptrace: true,
	//}
	//cmd.Stdout = os.Stdout
	runtime.LockOSThread()
	pinfo, err := os.StartProcess("/bin/echo", []string{"echo", "lee"}, &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout},
		Sys: &syscall.SysProcAttr{
			Ptrace: true,
		},
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	regs2 := &syscall.PtraceRegs{}
	fmt.Println(syscall.PtraceGetRegs(pinfo.Pid, regs2))
	ba := regs2.Rip
	//syscall.PtraceCont(pinfo.Pid, 0)
	//cpid := pinfo.Pid
	//syscall.PtraceDetach(pinfo.Pid)
	var wopt syscall.WaitStatus
	regs1 := &syscall.PtraceRegs{}
	for regs1 != nil && regs1.Orig_rax != 1 {
		//////////////////////////////////////////before syscall//////////
		syscall.PtraceSyscall(pinfo.Pid, 0)
		syscall.Wait4(pinfo.Pid, &wopt, 0, nil)
		fmt.Println(wopt)
		fmt.Println(syscall.PtraceGetRegs(pinfo.Pid, regs1))
		fmt.Println(regs1.Orig_rax)
		if regs1.Orig_rax == 1 {
			fmt.Printf("%#v\n", regs1)
			out := make([]byte, int(regs1.Rbx))
			syscall.PtracePeekData(pinfo.Pid, uintptr(regs1.Rsi), out)
			fmt.Println("Get data", string(out))
		}
		fmt.Println("line:", regs1.Rip-ba)
		syscall.PtraceSyscall(pinfo.Pid, 0)
		syscall.Wait4(pinfo.Pid, &wopt, 0, nil)
		fmt.Println(wopt)
		fmt.Println(syscall.PtraceGetRegs(pinfo.Pid, regs1))
		fmt.Println(regs1.Orig_rax)
		//////////////////////////////////////////////after syscall/////////
	}
	syscall.PtraceSyscall(pinfo.Pid, 0)
	syscall.Wait4(pinfo.Pid, &wopt, 0, nil)
	fmt.Println(wopt)
	fmt.Println(syscall.PtraceGetRegs(pinfo.Pid, regs1))
	fmt.Println(regs1.Orig_rax)
	time.Sleep(time.Second)
	return
	//syscall.PtraceAttach(cpid)
	//syscall.PtraceSingleStep(cpid)
	//cmd.Start()
	//cmd.Wait()
	//cpid := cmd.ProcessState.Pid()
	//var wopt syscall.WaitStatus
	fmt.Println(pinfo.Pid)
	time.Sleep(15 * time.Second)
	for {
		fmt.Printf("msg>>>")
		_, err := pinfo.Wait()
		if err != nil {
			fmt.Println(err)
			return
		}

		//fmt.Println(syscall.PtraceGetEventMsg(pinfo.Pid))
		//wpid, err := syscall.Wait4(cpid, &wopt, 0, nil)
		//if err != nil {
		//	fmt.Println(err)
		//	break
		//}
		////fmt.Println(cpid, wpid, wopt)
		//if wopt.Exited() {
		//	break
		//}
		regs := &syscall.PtraceRegs{}
		fmt.Println(syscall.PtraceGetRegs(pinfo.Pid, regs))
		fmt.Printf("%#v\n", regs)
		fmt.Println(regs.Orig_rax)

		//time.Sleep(time.Second)
		syscall.PtraceCont(pinfo.Pid, 59)
		fmt.Println(syscall.PtraceGetRegs(pinfo.Pid, regs))
	}

	time.Sleep(time.Second)
	fmt.Println("ssss")
}

