package main

import (
	"github.com/artpar/goms/server"
	"github.com/rcrowley/goagain"
	"log"
	//"os"
	"fmt"
	"net"
	//"sync"
	"syscall"
	"time"
	"github.com/GeertJohan/go.rice"
	"net/http"
	"sync"
)

func init() {
	goagain.Strategy = goagain.Double
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("pid:%d ", syscall.Getpid()))
}

func main() {

	boxStatic1, err := rice.FindBox("gomsweb/dist/static")
	log.Println("Failed to open dist/static: %v", err)
	boxRoot1, err := rice.FindBox("gomsweb/dist")
	log.Println("Failed to open dist: %v", err)

	var boxStatic, boxRoot http.FileSystem
	if err != nil {
		boxStatic = http.Dir("gomsweb/dist/static")
		boxRoot = http.Dir("gomsweb/dist")
	} else {
		boxStatic = boxStatic1.HTTPBox()
		boxRoot = boxRoot1.HTTPBox()
	}

	// Inherit a net.Listener from our parent process or listen anew.
	ch := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	l, err := goagain.Listener()
	if nil != err {

		// Listen on a TCP or a UNIX domain socket (TCP here).
		l, err = net.Listen("tcp", "127.0.0.1:48879")
		if nil != err {
			log.Fatalln(err)
		}
		log.Println("listening on", l.Addr())

		// Accept connections in a new goroutine.
		go server.Main(boxRoot, boxStatic)
		go serve(l, ch, wg)

	} else {

		// Resume listening and accepting connections in a new goroutine.
		log.Println("resuming listening on", l.Addr())
		go server.Main(boxRoot, boxStatic)
		go serve(l, ch, wg)

		// If this is the child, send the parent SIGUSR2.  If this is the
		// parent, send the child SIGQUIT.
		if err := goagain.Kill(); nil != err {
			log.Fatalln(err)
		}

	}

	// Block the main goroutine awaiting signals.
	sig, err := goagain.Wait(l)
	if nil != err {
		log.Fatalln(err)
	}

	// Do whatever's necessary to ensure a graceful exit like waiting for
	// goroutines to terminate or a channel to become closed.
	//
	// In this case, we'll close the channel to signal the goroutine to stop
	// accepting connections and wait for the goroutine to exit.
	close(ch)
	wg.Wait()

	// If we received SIGUSR2, re-exec the parent process.
	if goagain.SIGUSR2 == sig {
		if err := goagain.Exec(l); nil != err {
			log.Fatalln(err)
		}
	}
}

// A very rude server that says hello and then closes your connection.
func serve(l net.Listener, ch chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for {

		// Break out of the accept loop on the next iteration after the
		// process was signaled and our channel was closed.
		select {
		case <-ch:
			return
		default:
		}

		// Set a deadline so Accept doesn't block forever, which gives
		// us an opportunity to stop gracefully.
		l.(*net.TCPListener).SetDeadline(time.Now().Add(100e6))

		c, err := l.Accept()
		if nil != err {
			if goagain.IsErrClosing(err) {
				return
			}
			if err.(*net.OpError).Timeout() {
				continue
			}
			log.Fatalln(err)
		}
		c.Write([]byte("Hello, world!\n"))
		c.Close()
	}
}
