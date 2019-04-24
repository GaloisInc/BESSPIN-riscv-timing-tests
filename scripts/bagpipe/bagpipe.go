package bagpipe
// This is a copy of package gitlab.com/ashay/bagpipe
// with a small change to ExecCommand

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"runtime"
	"sync"
	"time"

	"io/ioutil"
	"os/exec"
	"os/user"
	"path/filepath"
)

type kernel_t func(bytes.Buffer) bytes.Buffer

type task_t struct {
	input  bytes.Buffer
	output bytes.Buffer
}

type Sprinter_t struct {
	tasks        chan task_t
	kernel_fn    kernel_t
	wait_group   sync.WaitGroup
	input_queue  chan bytes.Buffer
	channel_size int

	work_items int
}

func NewSprinter(kernel_fn kernel_t, pool_size int, channel_size int) *Sprinter_t {
	sprinter := new(Sprinter_t)

	sprinter.work_items = 0
	sprinter.kernel_fn = kernel_fn
	sprinter.channel_size = channel_size

	sprinter.input_queue = make(chan bytes.Buffer)
	sprinter.tasks = make(chan task_t, channel_size)

	// start worker threads.
	for worker := 0; worker < pool_size; worker += 1 {
		sprinter.wait_group.Add(1)

		go func(inputs chan bytes.Buffer, tasks chan task_t,
			wait_group *sync.WaitGroup) {
			defer wait_group.Done()

			for input := range inputs {
				task := task_t{input: input, output: sprinter.kernel_fn(input)}
				tasks <- task
			}
		}(sprinter.input_queue, sprinter.tasks, &sprinter.wait_group)
	}

	return sprinter
}

func (sprinter *Sprinter_t) FeedWorker(args ...interface{}) {
	for _, arg := range args {
		var encoded_output bytes.Buffer
		encoder := gob.NewEncoder(&encoded_output)

		err := encoder.Encode(arg)
		CheckError(err)

		if sprinter.work_items >= sprinter.channel_size {
			log.Fatal("Adding more work items than the channel size!")
		}

		sprinter.work_items += 1
		sprinter.input_queue <- encoded_output
	}
}

func (sprinter *Sprinter_t) CloseQueue() {
	close(sprinter.input_queue)
	sprinter.wait_group.Wait()
}

func (sprinter Sprinter_t) ResultCount() int {
	return sprinter.work_items
}

func (sprinter Sprinter_t) ReadResult() (bytes.Buffer, bytes.Buffer) {
	if sprinter.work_items == 0 {
		log.Fatal("Trying to read beyond the number of available results!")
	}

	sprinter.work_items -= 1
	task := <-sprinter.tasks

	return task.input, task.output
}

func CheckError(err error) {
	if err != nil {
		ClearStatus()
		_, function, line, _ := runtime.Caller(1)

		log.Printf("[error] %s:%d ", function, line)
		log.Fatal(err)
	}
}

func ExecCommand(cmd string, working_dir string) string {
	cmd_obj := exec.Command("sh", "-c", cmd)
	// cmd_obj.Stdin = os.Stdin
	// the above line breaks terminal emulators!
	cmd_obj.Dir = working_dir

	stdout, err := cmd_obj.CombinedOutput()

	if err != nil {
		ClearStatus()
		log.Printf("Failed to execute command:")
		log.Printf("cmd: " + cmd)
		log.Printf("dir: " + working_dir)
		log.Printf("out: " + string(stdout[:]))
		log.Fatal("terminating.")
	}

	return string(stdout[:])
}

func ExecCommandWithEnv(cmd string, working_dir string, vars []string) string {
	cmd_obj := exec.Command("sh", "-c", cmd)
	cmd_obj.Stdin = os.Stdin
	cmd_obj.Dir = working_dir
	cmd_obj.Env = append(os.Environ(), vars...)

	stdout, err := cmd_obj.CombinedOutput()

	if err != nil {
		ClearStatus()
		log.Printf("Failed to execute command:")
		log.Printf("cmd: " + cmd)
		log.Printf("dir: " + working_dir)
		log.Printf("out: " + string(stdout[:]))
		log.Fatal("terminating.")
	}

	return string(stdout[:])
}

func ReadFile(filename string) string {
	contents, err := ioutil.ReadFile(filename)
	CheckError(err)

	return string(contents[:])
}

func WriteFile(filename string, contents string) {
	byte_array := []byte(contents)

	file, err := os.Create(filename)
	CheckError(err)

	_, err = file.Write(byte_array)
	CheckError(err)

	file.Close()
}

func AppendFile(filename string, contents string) {
	byte_array := []byte(contents)

	mode := os.O_APPEND | os.O_WRONLY | os.O_CREATE
	file, err := os.OpenFile(filename, mode, 0644)
	CheckError(err)

	_, err = file.Write(byte_array)
	CheckError(err)

	file.Close()
}

func DeleteFile(filename string) {
	err := os.Remove(filename)
	if os.IsExist(err) || err != nil {
		log.Fatal(err)
	}
}

func MoveFile(old_path string, new_path string) {
	err := os.Rename(old_path, new_path)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateDirectory(dir string) {
	err := os.Mkdir(dir, 0755)
	CheckError(err)
}

func DeleteDirectory(dir string) {
	err := os.RemoveAll(dir)
	CheckError(err)
}

func CreateTempDirectory(prefix string) string {
	name, err := ioutil.TempDir("", prefix)
	CheckError(err)

	return name
}

func CreateTempFile(prefix string) string {
	tmp_file, err := ioutil.TempFile("", prefix)
	CheckError(err)

	filename := tmp_file.Name()

	tmp_file.Close()
	return filename
}

func ListFiles(dir string, regex regexp.Regexp) []string {
	dir_obj, err := os.Open(dir)
	CheckError(err)

	files, err := dir_obj.Readdir(0)
	CheckError(err)

	dir_obj.Close()

	var paths []string

	for _, file := range files {
		if regex.MatchString(file.Name()) {
			path, err := filepath.Abs(filepath.Dir(dir + "/" + file.Name()))
			CheckError(err)

			paths = append(paths, path+"/"+file.Name())
		}
	}

	return paths
}

func UpdateStatus(msg string) {
	fmt.Print("\r                                                            ")
	fmt.Print("\r" + msg)
}

func ClearStatus() {
	UpdateStatus("")
}

func WorkingDirectory() string {
	directory, err := os.Getwd()
	CheckError(err)

	return directory
}

func ScriptDirectory() string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		log.Fatal("Failed to obtain script directory.\n")
	}

	return path.Dir(filename)
}

func HomeDirectory() string {
	usr, err := user.Current()
	CheckError(err)

	return usr.HomeDir
}

func ModificationTime(filename string) time.Time {
	info, err := os.Stat(filename)
	CheckError(err)

	return info.ModTime()
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func Username() string {
	user, err := user.Current()
	CheckError(err)

	return user.Username
}

func GetFileSize(filename string) int64 {
	info, err := os.Stat(filename)
	CheckError(err)

	return info.Size()
}
