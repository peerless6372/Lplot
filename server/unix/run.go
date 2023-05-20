package unix

import (
	"net"
	"net/http"
	"os"

	"lzh/gin-gonic/gin"
)

const (
	socketPath = "/usr/local/var/run/"
	sockName   = "go.sock"
)

func Start(engine *gin.Engine) {
	var err error
	go func() {
		if _, err := os.Stat(socketPath); os.IsNotExist(err) {
			err = os.MkdirAll(socketPath, os.ModePerm)
			if err != nil {
				panic("mkdir " + socketPath + "error: " + err.Error())
			}
		}

		fd := socketPath + sockName
		if _, err = os.Stat(fd); err == nil {
			_ = os.Remove(fd)
		}

		listener, err := net.Listen("unix", fd)
		if err != nil {
			panic("listen unix error: " + err.Error())
		}

		defer listener.Close()
		defer os.Remove(fd)

		if err = os.Chmod(fd, 0666); err != nil {
			panic("unix socket chmod error: " + err.Error())
		}

		if err = http.Serve(listener, engine); err != nil {
			panic("runUnix error: " + err.Error())
		}
	}()
}
