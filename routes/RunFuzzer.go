package routes

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
)

func errReturn(c *gin.Context, err error) {
	c.JSON(500, gin.H{
		"message": err,
	})
}

func makeFile(ctx context.Context, c *gin.Context) error {
	cmd := exec.CommandContext(ctx, "make", "clean")
	output, err := cmd.CombinedOutput()
	log.Println(string(output))
	if err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "make")
	output, err = cmd.CombinedOutput()
	log.Println(string(output))
	if err != nil {
		return err
	}
	return nil
}

func RunFuzzer(apiGroup *gin.RouterGroup) {
	apiGroup.POST("/runAFLPlusPlus", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		currDir, err := os.Getwd()
		if err != nil {
			errReturn(c, err)
			return
		}
		err = os.Chdir(currDir + "/target/vul")
		if err != nil {
			errReturn(c, err)
			return
		}

		// MakeFile
		err = makeFile(ctx, c)
		if err != nil {
			errReturn(c, err)
			return
		}

		// Run File
		cmd := exec.CommandContext(ctx, "./build/vul")
		// cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout

		err = cmd.Start()
		if err != nil {
			errReturn(c, err)
			return
		}

		finish := make(chan error, 1)
		go func() {
			finish <- cmd.Wait()
		}()

		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				cmd.Process.Kill()
				c.JSON(200, gin.H{
					"message": "after 1 second, kill the process",
				})
				return
			}
		case err = <-finish:
			if err != nil {
				errReturn(c, err)
				return
			}
			c.JSON(200, gin.H{
				"message": "finish the task",
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "runFuzzer",
		})
	})
}
