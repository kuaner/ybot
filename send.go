package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var errAudio = errors.New("")

type audioMsg struct {
	chatID    int64
	title     string
	audio     string
	performer string
	thumb     string
	duration  int
}

// 由于使用的api库不支持sendAudio的时候传thumb，这里重写了一个
func sendAudio(bot *tgbotapi.BotAPI, msg audioMsg) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// set file
	if msg.audio != "" {
		if err := addFileToWriter(writer, "audio", msg.audio); err != nil {
			return err
		}
	} else {
		return errAudio
	}
	if msg.thumb != "" {
		if err := addFileToWriter(writer, "thumb", msg.thumb); err != nil {
			return err
		}
	}
	// set params
	if msg.chatID != 0 {
		if err := writer.WriteField("chat_id", fmt.Sprintf("%d", msg.chatID)); err != nil {
			return err
		}
	} else {
		return errAudio
	}
	if msg.title != "" {
		if err := writer.WriteField("title", msg.title); err != nil {
			return err
		}
	}
	if msg.performer != "" {
		if err := writer.WriteField("performer", msg.performer); err != nil {
			return err
		}
	}
	if msg.duration != 0 {
		if err := writer.WriteField("duration", fmt.Sprintf("%d", msg.duration)); err != nil {
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return err
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendAudio", bot.Token)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := bot.Client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusInternalServerError {
		return errHTTP
	}
	return nil
}

func addFileToWriter(writer *multipart.Writer, fieldName string, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	part, err := writer.CreateFormFile(fieldName, filepath.Base(file))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, f)
	return err
}
