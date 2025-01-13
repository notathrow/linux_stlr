package fileutil

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Copy(src, dst string) (err error) {
	file, err := os.Stat(src)
	if err != nil {
		return err
	}

	if file.IsDir() {
		err = CopyDir(src, dst)
	} else {
		err = CopyFile(src, dst)
	}

	return
}

func CopyFile(src, dst string) (err error) {
	c, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, err = f.Write(c)
	return err
}

func CopyDir(src, dst string) error {
	var err error
	topDir, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for i := 0; i < len(topDir); i++ {
		if !topDir[i].IsDir() {
			srcfile, err := os.ReadFile(src + string(os.PathSeparator) + topDir[i].Name())
			if err != nil {
				return err
			}

			dstfile, err := os.Create(dst + string(os.PathSeparator) + topDir[i].Name())
			if err != nil {
				return err
			}
			defer dstfile.Close()

			_, err = dstfile.Write(srcfile)
			if err != nil {
				return err
			}
		} else {
			srcchildfolder := src + string(os.PathSeparator) + topDir[i].Name()
			dstchildfolder := dst + string(os.PathSeparator) + topDir[i].Name()
			err = os.Mkdir(dstchildfolder, 0770)
			if err != nil {
				return err
			}

			err = CopyDir(srcchildfolder, dstchildfolder)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func ReadFile(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func ReadLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make([]string, 0)
	buf := bufio.NewReader(f)

	for {
		line, _, err := buf.ReadLine()
		l := string(line)
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		result = append(result, l)
	}

	return result, nil
}

// needs to have os compatible path separator or else won't work
func ZipDir(dir string) (string, error) {

	if !(strings.HasSuffix(dir, "\\") || strings.HasSuffix(dir, "/")) {
		dir = fmt.Sprintf("%s%c", dir, os.PathSeparator)
	}
	files, err := ListFiles(dir)
	if err != nil {
		return "", err
	}

	for i := 0; i < len(files); i++ {
		files[i] = strings.Replace(files[i], dir, "", 1)
	}
	tempfile := os.TempDir() + string(os.PathSeparator) + filepath.Base(dir) + ".zip"
	zippath, err := ZipFiles(tempfile, files, dir)
	if err != nil {
		return "", err
	}
	return zippath, nil
}

func ListFiles(dir string) ([]string, error) {
	var filepaths []string
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(files); i++ {
		if files[i].IsDir() {
			cfiles, err := ListFiles(filepath.Join(dir, files[i].Name()))
			if err != nil {
				return nil, err
			}
			for j := 0; j < len(cfiles); j++ {
				filepaths = append(filepaths, cfiles[j])
			}
		} else {
			filepaths = append(filepaths, filepath.Join(dir, files[i].Name()))
		}
	}
	return filepaths, nil
}

func ZipFiles(name string, files []string, dir string) (string, error) {
	zipfile, err := os.Create(name)
	if err != nil {
		return "", err
	}
	defer zipfile.Close()
	w := zip.NewWriter(zipfile)
	defer w.Close()
	for _, file := range files {
		filezippath := filepath.Base(dir) + string(os.PathSeparator) + file
		zf, err := w.Create(filezippath)
		if err != nil {
			return "", err
		}

		f, err := os.Open(dir + string(os.PathSeparator) + file)
		if err != nil {
			return "", err
		}

		io.Copy(zf, f)
		f.Close()
	}
	return name, nil
}
