package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func name(in string) (o1, o2 string) {
	md5Str := fmt.Sprintf("%x", md5.Sum([]byte(in)))
	return filepath.Join(config.outputPath, md5Str[:len(md5Str)/2]),
		filepath.Join(config.outputPath, md5Str[len(md5Str)/2:])

}

func clean(input ...string) {
	for _, i := range input {
		if i != "" {
			os.RemoveAll(i)
		}
	}
}

func ffmpegExist() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

func ffprobeExist() bool {
	_, err := exec.LookPath("ffprobe")
	return err == nil
}
