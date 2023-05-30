package routes

import (
	"context"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func errReturn(c *gin.Context, err error, errMsg string) {
	log.Println(err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"message": errMsg,
	})
}

func setUpTargetDir(workDir, targetDir, targetName string, file *multipart.FileHeader, c *gin.Context) error {
	os.MkdirAll(targetDir, 0777)
	os.MkdirAll(filepath.Join(targetDir, "src"), 0777)
	os.MkdirAll(filepath.Join(targetDir, "build"), 0777)
	os.MkdirAll(filepath.Join(targetDir, "test"), 0777)

	// Save C code
	err := c.SaveUploadedFile(file, filepath.Join(targetDir, "src", targetName+".c"))
	if err != nil {
		return err
	}

	// Copy Makefile
	cmd := exec.Command("cp", filepath.Join(workDir, "make", "Makefile"), filepath.Join(targetDir, "Makefile"))
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func makeFile(AFLDir, targetDir, targetName string) error {
	gccEnv := "CC=" + filepath.Join(AFLDir, "afl-gcc")
	AsanEnv := "AFL_USE_ASAN=1"

	cmd := exec.Command("make", "-C", targetDir, "clean")
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("make", "-C", targetDir, "target="+targetName)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, gccEnv, AsanEnv)

	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func createTestDir(targetDir, targetName, currentTime string) error {
	testDir := filepath.Join(targetDir, "test")
	os.MkdirAll(filepath.Join(testDir, currentTime), 0777)
	os.MkdirAll(filepath.Join(testDir, currentTime, "in"), 0777)
	os.MkdirAll(filepath.Join(testDir, currentTime, "out"), 0777)

	cmd := exec.Command("cp", filepath.Join(targetDir, "build", targetName), filepath.Join(testDir, currentTime, targetName))
	err := cmd.Run()
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(filepath.Join(testDir, currentTime, "in"))
	if err != nil {
		return err
	}
	if len(files) > 0 {
		return nil
	}

	file, err := os.Create(filepath.Join(testDir, currentTime, "in", "input"))
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
	cmd := exec.Command(filepath.Join(AFLDir, "afl-gotcpu"))
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
	files, err := ioutil.ReadDir(filepath.Join(targetDir, "test", currentTime, "out", "default", "crashes"))
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

func runFuzzer(targetDir, targetName, currentTime, AFLDir, avCpu string) error {
	inputDir := filepath.Join(targetDir, "test", currentTime, "in")
	outputDir := filepath.Join(targetDir, "test", currentTime, "out")
	execDir := filepath.Join(targetDir, "test", currentTime, targetName)

	screenCmd := exec.Command("screen", "-dmS", targetName+"-"+currentTime)
	err := screenCmd.Run()
	if err != nil {
		return err
	}

	aflCmdString := AFLDir + "/afl-fuzz -i " + inputDir + " -o " + outputDir + " -b " + avCpu + " -m none -- " + execDir + " @@"
	screenExecCmd := exec.Command("screen", "-S", targetName+"-"+currentTime, "-X", "stuff", "bash -c \""+aflCmdString+"\"\n")

	// *cmd = exec.CommandContext(*ctx, AFLDir+"/afl-fuzz", "-i", inputDir, "-o", outputDir,
	// 	"-b", avCpu, "-m", "none", "--", execDir, "@@")
	err = screenExecCmd.Run()
	if err != nil {
		return err
	}

	// go func() {
	// 	finish <- (*cmd).Wait()
	// }()

	return nil
}

type FileUploadFormData struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func RunFuzzer(apiGroup *gin.RouterGroup, workDir string) {
	var mutex sync.Mutex
	apiGroup.POST("/runAFLPlusPlus", func(c *gin.Context) {
		formData := &FileUploadFormData{}
		if err := c.ShouldBind(formData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		targetName := strings.Split(formData.File.Filename, ".")[0]
		targetDir := filepath.Join(workDir, "target", targetName)
		AFLDir := filepath.Join(workDir, "AFLplusplus")

		// Set up target dir and save c code
		setUpTargetDir(workDir, targetDir, targetName, formData.File, c)

		// MakeFile
		err := makeFile(AFLDir, targetDir, targetName)
		if err != nil {
			errReturn(c, err, "makefile error")
			return
		}
		log.Println("[*] Make: make" + targetName + ".c success")

		// Create Test dir
		currentTime := time.Now().Format("20060102_150405")
		err = createTestDir(targetDir, targetName, currentTime)
		if err != nil {
			errReturn(c, err, "create test dir error")
			return
		}

		// afl-gotcpu cannot run in same time
		mutex.Lock()
		// Get available cpu
		avCpu, err := getAvailableCpu(AFLDir)
		if err != nil {
			errReturn(c, err, "get available cpu error")
			return
		}
		if avCpu == "-1" {
			c.JSON(http.StatusNotFound, gin.H{"message": "no available cpu"})
			return
		}

		// Run fuzzer
		err = runFuzzer(targetDir, targetName, currentTime, AFLDir, avCpu)
		if err != nil {
			errReturn(c, err, "Run fuzzer error")
			return
		}
		log.Println("[*] Fuzzing: fuzzing " + targetName + ".c ...")
		log.Println("[*] Remote Screen: screen -r " + targetName + "-" + currentTime)

		// Set 3 min timeout
		// ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		time.Sleep(3 * time.Second)
		mutex.Unlock()

		// Timeout(3mins)
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			killScreenCmd := exec.Command("screen", "-S", targetName+"-"+currentTime, "-X", "quit")
			err = killScreenCmd.Run()
			if err != nil {
				errReturn(c, err, "Kill screen error")
				return
			}

			if checkCrash(targetDir, currentTime) {
				c.JSON(http.StatusOK, gin.H{"message": targetName + ".c : Crash Found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": targetName + ".c : Crash Not Found"})
			return
		}

		c.JSON(400, gin.H{
			"message": "Bad Request",
		})
	})

	// Get gpu info api
	// apiGroup.GET("/gotcpu", func(c *gin.Context) {
	// 	err := os.Chdir(workDir + "/AFLplusplus")
	// 	if err != nil {
	// 		errReturn(c, err, "change dir error")
	// 		return
	// 	}
	// 	log.Println(workDir)

	// 	cmd := exec.Command("./afl-gotcpu")
	// 	output, err := cmd.CombinedOutput()
	// 	if err != nil {
	// 		errReturn(c, err, "exec afl-gotcpu fail")
	// 		return
	// 	}

	// 	cpuInfo := strings.ReplaceAll(string(output), "\r", "")
	// 	lines := strings.Split(cpuInfo, "\n")

	// 	for _, line := range lines {
	// 		fields := strings.FieldsFunc(line, func(r rune) bool {
	// 			return r == '#' || r == ':' || r == ' ' || r == '('
	// 		})
	// 		if len(fields) == 0 {
	// 			continue
	// 		}
	// 		if fields[0] == "Core" {
	// 			if fields[2] == "OVERBOOKED" {
	// 				continue
	// 			}
	// 			c.JSON(200, gin.H{
	// 				"message": fields,
	// 			})
	// 			return
	// 		}
	// 	}

	// 	c.JSON(200, gin.H{
	// 		"message": "OK",
	// 	})
	// })

	// apiGroup.POST("/uploadFile", func(c *gin.Context) {
	// 	// Max 32 MB
	// 	err := c.Request.ParseMultipartForm(32 << 20)
	// 	if err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{
	// 			"message": err.Error(),
	// 		})
	// 		return
	// 	}

	// 	formData := &FileUploadFormData{}
	// 	if err := c.ShouldBind(formData); err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{
	// 			"message": err.Error(),
	// 		})
	// 		return
	// 	}

	// 	filename := formData.File.Filename
	// 	savePath := filepath.Join(workDir, filename)

	// 	err = c.SaveUploadedFile(formData.File, savePath)
	// 	if err != nil {
	// 		errReturn(c, err, "save file error")
	// 		return
	// 	}

	// 	c.JSON(200, gin.H{
	// 		"message": "OK",
	// 	})

	// })
}
