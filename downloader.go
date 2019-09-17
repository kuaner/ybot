package main

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"
)

var client = http.Client{Timeout: time.Second * 180}
var (
	errHTTP     = errors.New("Http error")
	errWrite    = errors.New("Write error")
	errDownload = errors.New("Download error")
)

//每个线程下载文件的大小 4M
var packageSize int64 = 1048576 * 4

func download(url, cachePath string, size int64) error {
	var localFileSize int64
	var file *os.File
	if info, e := os.Stat(cachePath); e != nil {
		if os.IsNotExist(e) {
			if createFile, err := os.Create(cachePath); err == nil {
				file = createFile
			} else {
				return err
			}
		} else {
			return e
		}
	} else {
		localFileSize = info.Size()
	}
	if localFileSize == size {
		return nil
	}
	return dispSliceDownload(file, size, url)
}

func dispSliceDownload(file *os.File, ContentLength int64, url string) error {
	threadGroup, _ := errgroup.WithContext(context.Background())
	// threadGroup := sync.WaitGroup{}
	defer file.Close()
	//文件总大小除以 每个线程下载的大小
	i := ContentLength / packageSize
	//保证文件下载完整
	if ContentLength%packageSize > 0 {
		i++
	}
	//分配下载线程
	for count := 0; count < int(i); count++ {
		//计算每个线程下载的区间,起始位置
		var start int64
		var end int64
		start = int64(int64(count) * packageSize)
		end = start + packageSize
		if int64(end) > ContentLength {
			end = end - (end - ContentLength)
		}
		//构建请求
		if req, e := http.NewRequest("GET", url, nil); e == nil {
			req.Header.Set(
				"Range",
				"bytes="+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10))
			// call func
			threadGroup.Go(func() error {
				return sliceDownload(req, file, start)
			})
		} else {
			return errHTTP
		}
	}
	//等待所有线程完成下载
	return threadGroup.Wait()
}

func sliceDownload(request *http.Request, file *os.File, start int64) error {
	if response, e := client.Do(request); e == nil && response.StatusCode == 206 {
		defer response.Body.Close()
		if bytes, i := ioutil.ReadAll(response.Body); i == nil {
			//从我们计算好的起点写入文件
			_, err := file.WriteAt(bytes, start)
			if err != nil {
				return errWrite
			}
			return nil
		}
		return errDownload
	}
	return errDownload
}

func downloadThumb(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
