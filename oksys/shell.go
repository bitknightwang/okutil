package oksys

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bitknightwang/okutil/oklog"
)

func Contains(list []string, element string) bool {
	for _, el := range list {
		if el == element {
			return true
		}
	}

	return false
}

func IsEmpty(text string) bool {
	return strings.TrimSpace(text) == ""
}

func IsNotEmpty(text string) bool {
	return !IsEmpty(text)
}

func RunShellCommand(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).Output()

	if err != nil {
		output := fmt.Sprintf("Error on running %v %v", cmd, args)
		oklog.Errorf(output+"\n%v", err)
		return output, err
	}

	output := string(out)
	oklog.Debugf("Command: %s %v\nOutput:\n%s", cmd, args, output)
	// return output, nil
	return strings.TrimSpace(strings.TrimRight(output, "\n")), nil
}

func MakeDirs(path string, mode os.FileMode) error {
	if len(path) == 0 {
		oklog.Errorf("Illegal path %v", path)
		return errors.New("empty path")
	}

	if Exists(path) {
		// do nothing
		return nil
	}

	if mode == 0 {
		mode = 0755
	}

	if err := os.MkdirAll(path, mode); err != nil {
		oklog.Errorf("Failed to create dir %v\n%v", path, err)
		return err
	}
	return nil
}

func Rename(before, after string) error {
	if len(before) == 0 || len(after) == 0 {
		errMsg := fmt.Sprintf("Illegal input before:%v after:%v", before, after)
		oklog.Error(errMsg)
		return errors.New(errMsg)
	}

	if err := os.Rename(before, after); err != nil {
		oklog.Errorf("Failed to rename %v to %v\n%v", before, after, err)
		return err
	}

	return nil
}

func Delete(path string) error {
	if len(path) == 0 {
		oklog.Errorf("Illegal path %v", path)
		return errors.New("Empty path")
	}

	if err := os.RemoveAll(path); err != nil {
		oklog.Errorf("Failed to delete %v\n%v", path, err)
		return err
	}
	return nil
}

func IsDir(path string) bool {
	if IsEmpty(path) {
		return false
	}

	fi, err := os.Stat(path)
	if err != nil {
		oklog.Errorf("%v\n", err)
		return false
	}
	return fi.Mode().IsDir()
}

func IsFile(path string) bool {
	if IsEmpty(path) {
		return false
	}

	fi, err := os.Stat(path)
	if err != nil {
		oklog.Errorf("%v\n", err)
		return false
	}
	return fi.Mode().IsRegular()
}

func Exists(path string) bool {
	if IsEmpty(path) {
		return false
	}
	_, err := os.Stat(path)
	if err == nil || os.IsExist(err) {
		return true
	}

	return false
}

func DirExists(path string) bool {
	if !Exists(path) {
		return false
	}

	return IsDir(path)
}

func FileExists(path string) bool {
	if !Exists(path) {
		return false
	}

	return IsFile(path)
}

func EmptyDir(path string) error {
	if !DirExists(path) {
		// do nothing
		return nil
	}

	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(path, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func SoftLink(source, target string, force bool) error {
	if len(source) == 0 || len(target) == 0 {
		return fmt.Errorf("empty arguments! source: %v target:%v", source, target)
	}

	// check exists
	if Exists(target) {
		if force {
			if err := os.Remove(target); err != nil {
				return fmt.Errorf("failed to unlink: %v", err)
			}
		} else {
			oklog.Warnf("overwrite existing symlink %v", target)
		}
	}

	if err := os.Symlink(source, target); err != nil {
		return fmt.Errorf("Failed to create soft link %v -> %v\n%v", source, target, err)
	}

	return nil
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) error {

	// zip -r ../okj-account-fe-me-2021-01-12-01.zip .
	// can not include .. in zip file for security reasons.
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func SearchFiles(pattern string) ([]string, error) {
	if len(pattern) == 0 {
		return nil, fmt.Errorf("empty argument")
	}

	matchedFiles, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	return matchedFiles, nil
}
