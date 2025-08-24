// utility/fs_copy.go
package Utility

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Copy copies src file to dst, overwriting dst if it exists.
func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

// CopyFile copies one file to another using `cp` command.
func CopyFile(source string, dest string) (err error) {
	cmd := exec.Command("cp", source, dest)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
	}
	return err
}

// CopyDir recursively copies one directory to another using `cp -R`.
func CopyDir(source string, dest string) (err error) {
	CreateDirIfNotExist(dest)
	cmd := exec.Command("cp", "-R", source, dest)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
	} else {
		fmt.Println("Result: " + out.String())
	}
	return err
}

// Move copies and removes a file or directory. Uses rsync/mv depending on OS.
func Move(source string, dest string) (err error) {
	CreateDirIfNotExist(dest)
	var out, stderr bytes.Buffer

	if runtime.GOOS == "windows" {
		rsync := exec.Command("mv", source, dest)
		rsync.Stdout = &out
		rsync.Stderr = &stderr
		err = rsync.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			return
		}
	} else {
		rsync := exec.Command("rsync", "-a", source, dest)
		rsync.Stdout = &out
		rsync.Stderr = &stderr
		err = rsync.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			return
		}
	}

	rm := exec.Command("rm", "-rf", source)
	rm.Stdout = &out
	rm.Stderr = &stderr
	err = rm.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	fmt.Println("Result: " + out.String())
	return nil
}

// MoveFile copies a file to destination then deletes the original.
func MoveFile(source, destination string) (err error) {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	fi, err := src.Stat()
	if err != nil {
		return err
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	perm := fi.Mode() & os.ModePerm
	dst, err := os.OpenFile(destination, flag, perm)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		dst.Close()
		os.Remove(destination)
		return err
	}
	if err = dst.Close(); err != nil {
		return err
	}
	if err = src.Close(); err != nil {
		return err
	}
	if err = os.Remove(source); err != nil {
		return err
	}
	return nil
}

// CompressDir compresses a directory into a .tar.gz written to buf.
func CompressDir(src string, buf io.Writer) (int, error) {
	src = strings.ReplaceAll(src, "\\", "/")
	tmp := RandomUUID() + ".tar.gz"
	defer os.Remove(tmp)

	args := []string{"-czvf", tmp, "-C", src, "."}
	cmd := exec.Command("tar", args...)
	cmd.Dir = os.TempDir()

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("tar", "-czvf", tmp, "-C", src, ".")
		fmt.Println("fail to compress file with error: ", fmt.Sprint(err)+": "+stderr.String())
		return -1, err
	}

	data, err := ioutil.ReadFile(filepath.Join(os.TempDir(), tmp))
	if err != nil {
		return -1, err
	}
	buf.Write(data)
	return len(data), nil
}

// ExtractTarGz extracts a tar.gz archive and returns the path to the extracted dir.
func ExtractTarGz(r io.Reader) (string, error) {
	tmpDir := strings.ReplaceAll(os.TempDir(), "\\", "/")

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	archive := RandomUUID() + ".tar.gz"
	err = ioutil.WriteFile(filepath.Join(tmpDir, archive), buf, 0777)
	if err != nil {
		return "", err
	}

	output := filepath.Join(tmpDir, RandomUUID())
	CreateDirIfNotExist(output)

	wait := make(chan error)
	args := []string{"-xvzf", archive, "-C", output, "--strip-components", "1"}
	RunCmd("tar", tmpDir, args, wait)

	if err = <-wait; err != nil {
		fmt.Println("fail to run: tar ", args)
		return "", err
	}
	fmt.Println("archive is extracted at ", output, err)
	return output, nil
}

