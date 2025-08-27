// utility/file.go
package Utility

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Exists reports whether the named file or directory exists.
func Exists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// IsEmpty reports whether a directory is empty.
func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// ReadDir returns a sorted list of FileInfo for the specified directory.
func ReadDir(dirname string) ([]os.FileInfo, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name() < list[j].Name() })
	return list, nil
}

// CreateIfNotExists creates a directory with the given permissions if it doesn't already exist.
func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}
	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}
	return nil
}

// CreateDirIfNotExist creates a directory hierarchy (0755) if it doesn't exist.
func CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveDirContents deletes all children of a directory without removing the directory itself.
func RemoveDirContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	return nil
}

// RemoveContents is an alias for RemoveDirContents.
func RemoveContents(dir string) error {
	return RemoveDirContents(dir)
}

// FindFileByName recursively finds files by exact (or dotted-suffix) name.
func FindFileByName(path string, name string) ([]string, error) {
	path = strings.ReplaceAll(path, "\\", "/")
	files := make([]string, 0)
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasPrefix(name, ".") {
			if strings.HasSuffix(info.Name(), name) {
				files = append(files, strings.ReplaceAll(p, "\\", "/"))
			}
		} else if info.Name() == name {
			files = append(files, strings.ReplaceAll(p, "\\", "/"))
		}
		return nil
	})
	return files, err
}

// GetFileContentType attempts to sniff the content type from the first 512 bytes.
func GetFileContentType(out *os.File) (string, error) {
	buffer := make([]byte, 512)
	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(buffer)
	return contentType, nil
}

// GetFilePathsByExtension recursively collects files with the given extension under path.
func GetFilePathsByExtension(path string, extension string) []string {
	files, err := ioutil.ReadDir(path)
	results := make([]string, 0)
	if err == nil {
		for i := 0; i < len(files); i++ {
			if files[i].IsDir() {
				results = append(results, GetFilePathsByExtension(path+"/"+files[i].Name(), extension)...)
			} else if strings.HasSuffix(files[i].Name(), extension) {
				results = append(results, path+"/"+files[i].Name())
			}
		}
	}
	return results
}

// WriteStringToFile creates (or truncates) a file and writes the provided string.
func WriteStringToFile(filepath, s string) error {
	fo, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer fo.Close()

	_, err = io.Copy(fo, strings.NewReader(s))
	if err != nil {
		return err
	}
	return nil
}

// copyFileContents copies src to dst, overwriting dst if it exists.
func copyFileContents(src, dst string) (err error) {
	src = strings.ReplaceAll(src, "\\", "/")
	dst = strings.ReplaceAll(dst, "\\", "/")
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// CopySymLink recreates a symlink at dest pointing to the same target as source.
func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}

// GetExecName returns the executable name (without extension) from a path.
func GetExecName(path string) string {
	var startIndex, endIndex int
	startIndex = strings.LastIndex(path, string(os.PathSeparator))
	if startIndex > -1 {
		path = path[startIndex+1:]
	}
	endIndex = strings.LastIndex(path, ".")
	if endIndex > -1 && startIndex > -1 {
		path = path[:endIndex]
	}
	return path
}

// FileLine returns "file.go:line" for the caller, useful for diagnostics.
func FileLine() string {
	_, fileName, fileLine, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", fileName, fileLine)
}

// FunctionName returns the current function name, useful for diagnostics.
func FunctionName() string {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}

// DownloadFile fetches a remote URL and writes it to fileName.
func DownloadFile(URL, fileName string) error {
	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("received non 200 response code")
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// JsonErrorStr marshals a simple error descriptor (kept here for convenience).
func JsonErrorStr(functionName string, fileLine string, err error) string {
	str, _ := json.Marshal(map[string]string{
		"FunctionName": functionName,
		"FileLine":     fileLine,
		"ErrorMsg":     err.Error(),
	})
	return string(str)
}

/**
 * Read movie file metadata...
 */
func ReadMetadata(path string) (map[string]interface{}, error) {
	cmd := exec.Command(`ffprobe`, `-hide_banner`, `-loglevel`, `fatal`, `-show_format`, `-print_format`, `json`, `-i`, path)
	cmd.Dir = os.TempDir()

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		return nil, err
	}

	infos := make(map[string]interface{})
	err = json.Unmarshal(out.Bytes(), &infos)
	if err != nil {
		return nil, err
	}

	return infos, nil
}

/**
 * Store meta data into a file.
 */
func SetMetadata(path, key, value string) error {

	// ffmpeg -i input.mp4 -metadata title="The video titile" -c copy output.mp4
	path = strings.ReplaceAll(path, "\\", "/")
	ext := path[strings.LastIndex(path, ".")+1:]

	// ffmpeg -i input.mp4 -metadata title="The video titile" -c copy output.mp4
	// Try more than once...
	nbTry := 30
	var err error

	// Generate the video in a temp file...
	dest := strings.ReplaceAll(path, "."+ext, ".temp."+ext)
	if Exists(dest) {
		os.Remove(dest)
	}

	for nbTry > 0 {
		// Generate the video in a temp file...
		dest := strings.ReplaceAll(path, "."+ext, ".temp."+ext)
		if Exists(dest) {
			os.Remove(dest)
		}

		args := []string{"-i", path, "-c:v", "copy"}
		args = append(args, "-c:a", "copy", "-c:s", "mov_text", "-map", "0")
		args = append(args, `-metadata`, key+`=`+value, dest)

		wait := make(chan error)
		RunCmd("ffmpeg", filepath.Dir(path), args, wait)
		err = <-wait

		if err != nil || !Exists(dest) {
			fmt.Println("fail to create metadata with error ", err, " try again in 2 sec...", path, nbTry)
			nbTry-- // give it time
			time.Sleep(2 * time.Second)
		} else if Exists(dest) {
			// Remove the original file...
			err = os.Remove(path)
			if err != nil {
				return err
			}

			// rename the file...
			err = os.Rename(dest, path)
			if err != nil {
				return err
			}

			return nil
		}
		if err != nil {
			fmt.Println("fail to run command ", err)
			return err
		}

	}

	return err
}
