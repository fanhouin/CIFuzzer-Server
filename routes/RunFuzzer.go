package routes

import (
	"context"
	"fmt"
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

func makeFile(ctx context.Context, c *gin.Context, workDir string) error {
	gccEnv := "CC=" + workDir + "/AFLplusplus/afl-gcc"
	AsanEnv := "AFL_USE_ASAN=1"

	cmd := exec.CommandContext(ctx, "make", "clean")
	output, err := cmd.CombinedOutput()
	log.Println(string(output))
	if err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "make")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, gccEnv, AsanEnv)

	output, err = cmd.CombinedOutput()
	log.Println(string(output))
	if err != nil {
		return err
	}
	return nil
}

func RunFuzzer(apiGroup *gin.RouterGroup, workDir string) {
	apiGroup.POST("/runAFLPlusPlus", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		err := os.Chdir(workDir + "/target/vuln1")
		if err != nil {
			errReturn(c, err)
			return
		}

		// MakeFile
		err = makeFile(ctx, c, workDir)
		if err != nil {
			errReturn(c, err)
			return
		}

		// Run File
		cmd := exec.CommandContext(ctx, "./build/vuln1")
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
				fmt.Println("exec fail")
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

	apiGroup.GET("/gotcpu", func(c *gin.Context) {
		err := os.Chdir(workDir + "/AFLplusplus")
		if err != nil {
			errReturn(c, err)
			return
		}

		cmd := exec.Command("afl-gotcpu")
		output, err := cmd.CombinedOutput()
		if err != nil {
			errReturn(c, err)
			return
		}
		fmt.Println(string(output))
		c.JSON(200, gin.H{
			"message": "OK",
		})
	})
}
