package routes

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func errReturn(c *gin.Context, err error) {
	c.JSON(500, gin.H{
		"message": err,
	})
}

func makeFile(AFLDir string) error {
	gccEnv := "CC=" + AFLDir + "/afl-gcc"
	AsanEnv := "AFL_USE_ASAN=1"

	cmd := exec.Command("make", "clean")
	output, err := cmd.CombinedOutput()
	log.Println(string(output))
	if err != nil {
		return err
	}

	cmd = exec.Command("make")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, gccEnv, AsanEnv)

	output, err = cmd.CombinedOutput()
	log.Println(string(output))
	if err != nil {
		return err
	}
	return nil
}

func createTestDir(targetDir, targetName, currentTime string) error {
	testDir := targetDir + "/test/"
	os.MkdirAll(testDir, 0777)
	os.MkdirAll(testDir+"/"+currentTime, 0777)
	os.MkdirAll(testDir+"/"+currentTime+"/in", 0777)
	os.MkdirAll(testDir+"/"+currentTime+"/out", 0777)

	cmd := exec.Command("cp", targetDir+"/build/"+targetName, testDir+"/"+currentTime+"/"+targetName)
	err := cmd.Run()
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(testDir + "/" + currentTime + "/in")
	if err != nil {
		return err
	}
	if len(files) > 0 {
		return nil
	}

	file, err := os.Create(testDir + "/" + currentTime + "/in/input")
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString("test")
	if err != nil {
		return err
	}

	return nil
}

func getAvailableCpu(AFLDir string) (string, error) {
	cmd := exec.Command(AFLDir + "/afl-gotcpu")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "-1", err
	}

	cpuInfo := strings.ReplaceAll(string(output), "\r", "")
	lines := strings.Split(cpuInfo, "\n")

	for _, line := range lines {
		fields := strings.FieldsFunc(line, func(r rune) bool {
			return r == '#' || r == ':' || r == ' ' || r == '('
		})
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "Core" {
			if fields[2] == "OVERBOOKED" {
				continue
			}
			return fields[1], nil
		}
	}
	return "-1", nil
}

func checkCrash(targetDir, currentTime string) bool {
	files, err := ioutil.ReadDir(targetDir + "/test/" + currentTime + "/out/default/crashes")
	if err != nil {
		return false
	}
	if len(files) == 0 {
		return false
	}
	if len(files) == 1 && files[0].Name() == "README.txt" {
		return false
	}
	for _, file := range files {
		if strings.Contains(file.Name(), "id:") {
			return true
		}
	}
	return false
}

func RunFuzzer(apiGroup *gin.RouterGroup, workDir string) {
	apiGroup.POST("/runAFLPlusPlus", func(c *gin.Context) {
		targetName := c.PostForm("target")
		if targetName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "target is empty",
			})
			return
		}
		AFLDir := workDir + "/AFLplusplus"
		targetDir := workDir + "/target/" + targetName
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "target not found",
			})
			return
		}

		// Set 3 min timeout
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Set target dir
		err := os.Chdir(targetDir)
		if err != nil {
			errReturn(c, err)
			return
		}

		// MakeFile
		err = makeFile(AFLDir)
		if err != nil {
			errReturn(c, err)
			return
		}

		// Create Test dir
		currentTime := time.Now().Format("20060102_150405")
		err = createTestDir(targetDir, targetName, currentTime)
		if err != nil {
			errReturn(c, err)
			return
		}

		// Get available cpu
		avCpu, err := getAvailableCpu(AFLDir)
		if err != nil {
			errReturn(c, err)
			return
		}
		if avCpu == "-1" {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "no available cpu",
			})
			return
		}

		inputDir := targetDir + "/test/" + currentTime + "/in"
		outputDir := targetDir + "/test/" + currentTime + "/out"
		execDir := targetDir + "/test/" + currentTime + "/" + targetName
		cmd := exec.CommandContext(ctx, AFLDir+"/afl-fuzz", "-i", inputDir, "-o", outputDir,
			"-b", avCpu, "-m", "none", "--", execDir, "@@")
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
		// Timeout(3mins)
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				cmd.Process.Kill()
				if checkCrash(targetDir, currentTime) {
					c.JSON(200, gin.H{
						"message": "Crash Found",
					})
					return
				}
				c.JSON(200, gin.H{
					"message": "Crash Not Found",
					// "message": "After 3 Minutes, kill the process",
				})
				return
			}
		// Finish the process
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

		c.JSON(400, gin.H{
			"message": "Bad Request",
		})
	})

	// Get gpu info api
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

		cpuInfo := strings.ReplaceAll(string(output), "\r", "")
		lines := strings.Split(cpuInfo, "\n")

		for _, line := range lines {
			fields := strings.FieldsFunc(line, func(r rune) bool {
				return r == '#' || r == ':' || r == ' ' || r == '('
			})
			if len(fields) == 0 {
				continue
			}
			if fields[0] == "Core" {
				if fields[2] == "OVERBOOKED" {
					continue
				}
				c.JSON(200, gin.H{
					"message": fields,
				})
				return
			}
		}

		c.JSON(200, gin.H{
			"message": "OK",
		})
	})
}
