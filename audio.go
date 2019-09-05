package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cast"
)

const splitDuration = "6000" // 100 * 60 second

func cvt(input, output string) error {
	ctx, cfn := context.WithTimeout(context.Background(), time.Minute*30)
	defer cfn()
	// 这里不知道为什么用acc编码telegram会有限制，不足50M也报错
	// 如果是aac编码，转码速度会大幅提升
	args := []string{
		"-i",
		input,
		"-nostats",
		"-loglevel",
		"0",
		"-c:a",
		"libmp3lame", //aac
		"-b:a",
		"64k",
		"-f",
		"segment",
		"-segment_time",
		splitDuration,
		"-y",
		fmt.Sprintf("%s_%%02d.mp3", output), //m4a
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	return cmd.Run()
}

func duration(input string) int {
	// ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1
	ctx, cfn := context.WithTimeout(context.Background(), time.Second*5)
	defer cfn()
	args := []string{
		"-v",
		"error",
		"-show_entries",
		"format=duration",
		"-of",
		"default=noprint_wrappers=1:nokey=1",
		input,
	}
	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	d, err := cmd.Output()
	if err == nil {
		return cast.ToInt(strings.Split(string(d), ".")[0])
	}
	return 0
}
