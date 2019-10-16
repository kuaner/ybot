package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/iawia002/annie/downloader"
	"github.com/iawia002/annie/extractors/youtube"
	"github.com/iawia002/annie/request"
	"github.com/iawia002/annie/utils"
)

type task struct {
	title  string
	url    string
	size   int64
	chatID int64
	msgID  int
	yURL   string
}

const telegramFileSizeLimit = 50 * 1024 * 1024 // 50M

func startBot(updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI) {
	taskC := make(chan task, config.threadNumber)
	for i := 0; i < config.threadNumber; i++ {
		go process(taskC, bot)
	}
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		t := extract(update.Message.Text)
		if t.size == 0 {
			continue
		}
		log.Printf("Receive msg from %s %d %s", update.Message.From.UserName, update.Message.Chat.ID, update.Message.Text)
		txt := fmt.Sprintf(`<a href="%s">🔊</a> %s`, t.url, t.title)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
		msg.ParseMode = tgbotapi.ModeHTML
		// msg.ReplyToMessageID = update.Message.MessageID
		resp, err := bot.Send(msg)
		if err != nil {
			//TODO: backoff retry?
			continue
		}
		// delete message
		bot.DeleteMessage(tgbotapi.DeleteMessageConfig{
			ChatID:    update.Message.Chat.ID,
			MessageID: update.Message.MessageID,
		})
		if !config.hasFfmpeg {
			continue
		}
		t.chatID = update.Message.Chat.ID
		t.msgID = resp.MessageID
		select {
		case taskC <- t:
			// do nothing
		default:
			// send failed msg
			txt := msg.Text + "队列拥堵，将不会发送离线音频文件!"
			editMsg := tgbotapi.NewEditMessageText(t.chatID, t.msgID, txt)
			editMsg.ParseMode = tgbotapi.ModeHTML
			bot.Send(editMsg)
		}
	}
}

func process(taskC <-chan task, bot *tgbotapi.BotAPI) {
	for t := range taskC {
		input, output := name(t.yURL)
		m4a := input + ".m4a"
		log.Printf("Start download m4a for %s", t.yURL)
		err := download(t.url, m4a, t.size)
		if err != nil {
			log.Printf("Download m4a for %s failed", t.yURL)
			clean(m4a)
			continue
		}
		log.Printf("Download m4a for %s finished", t.yURL)

		log.Printf("Start convert m4a for %s", t.yURL)
		err = cvt(m4a, output)
		if err != nil {
			log.Printf("Convert m4a for %s failed", t.yURL)
			clean(m4a)
			clean(fileList(output)...)
			continue
		}
		log.Printf("Convert m4a for %s finished", t.yURL)
		author, thumb := metadata(t.yURL)
		if downloadThumb(thumb, output+".jpg") == nil {
			thumb = output + ".jpg"
		} else {
			thumb = ""
		}
		l := fileList(output)
		// telegram的播放列表，是后收到的会在上，所以这里要倒序发msg
		for idx := len(l) - 1; idx >= 0; idx-- {
			f := l[idx]
			msg := audioMsg{
				audio:  f,
				chatID: t.chatID,
			}
			if len(l) == 1 {
				msg.title = t.title
			} else {
				msg.title = fmt.Sprintf("(%d/%d) - %s", idx+1, len(l), t.title)
			}
			msg.performer = author
			msg.thumb = thumb
			msg.duration = duration(f)
			log.Printf("Send %s %s", f, msg.title)
			sendAudio(bot, msg)
		}
		clean(m4a, thumb)
		clean(l...)
		// delete message
		bot.DeleteMessage(tgbotapi.DeleteMessageConfig{
			ChatID:    t.chatID,
			MessageID: t.msgID,
		})
	}
}

func fileList(leading string) []string {
	l, err := filepath.Glob(fmt.Sprintf("%s_*.mp3", leading))
	if err != nil {
		return []string{}
	}
	return l
}

func extract(videoURL string) (t task) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from extract", r)
		}
	}()
	var (
		domain string
		err    error
		data   []downloader.Data
	)
	u, err := url.ParseRequestURI(videoURL)
	if err != nil {
		return
	}
	domain = utils.Domain(u.Host)
	switch domain {
	case "youtube", "youtu": // youtu.be
		data, err = youtube.Extract(videoURL)
	default:
		return
	}
	if err != nil {
		return
	}
	for _, item := range data {
		if item.Err != nil {
			// if this error occurs, the preparation step is normal, but the data extraction is wrong.
			// the data is an empty struct.
			continue
		}
		for _, x := range item.Streams {
			for _, v := range x.URLs {
				if v.Ext == "m4a" {
					t.title = item.Title
					t.url = v.URL
					t.size = v.Size
					t.yURL = videoURL
					return
				}
			}
		}
	}
	return
}

// 为了不修改annie的代码，只能在这里多获取一次
// 获取youtube的作者
func metadata(videoURL string) (string, string) {
	html, err := request.Get(videoURL, "https://www.youtube.com", nil)
	if err != nil {
		return "Youtube Red", ""
	}
	ytplayer := utils.MatchOneOf(html, `;ytplayer\.config\s*=\s*({.+?});`)[1]
	var youtube = struct {
		Args struct {
			PlayerResponse string `json:"player_response"`
		} `json:"args"`
	}{}
	json.Unmarshal([]byte(ytplayer), &youtube)
	author := utils.GetStringFromJson(youtube.Args.PlayerResponse, "videoDetails.author")
	if author == "" {
		return "Youtube Red", ""
	}
	videoid := utils.GetStringFromJson(youtube.Args.PlayerResponse, "videoDetails.videoId")
	if videoid == "" {
		return "Youtube Red", ""
	}
	// telegram 会自动裁剪，这里传高清封面
	thumb := fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", videoid)
	return author, thumb
}
