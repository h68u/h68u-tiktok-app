package util

import (
	"bufio"
	"github.com/gin-gonic/gin"
	"os/exec"
	"strings"
)

func ExecuteCmd(command string, c *gin.Context) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	outP, _ := cmd.StdoutPipe()
	errP, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		return "", err
	}
	ch := make(chan string, 100)
	stdoutScan := bufio.NewScanner(outP)
	stderrScan := bufio.NewScanner(errP)
	go func() {
		for stdoutScan.Scan() {
			line := stdoutScan.Text()
			ch <- line
		}
	}()
	go func() {
		for stderrScan.Scan() {
			line := stderrScan.Text()
			ch <- line
		}
	}()
	go func() {
		err = cmd.Wait()
		if err != nil && !strings.Contains(err.Error(), "exit status") {
			c.JSON(200, gin.H{
				"error": err.Error(),
			})
		}
		close(ch)
	}()
	var res string
	for line := range ch {
		res += line
	}
	return res, nil
}
