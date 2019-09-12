package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/kuaner/telegram-bot-api"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
)

func start() {
	bot, err := tgbotapi.NewBotAPI(config.token)
	if err != nil {
		log.Printf("Init Bot error %s", err.Error())
		return
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// in case of webhook issue
	bot.RemoveWebhook()

	if config.hook {
		_, err = bot.SetWebhook(tgbotapi.NewWebhook(fmt.Sprintf("https://%s/ybot/%s", config.domain, bot.Token)))
		if err != nil {
			log.Printf("Set webhook error %s", err.Error())
			return
		}
		updates := bot.ListenForWebhook("/ybot/" + bot.Token)
		if config.acme {
			certManager := autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist(config.domain), //your domain here
				Cache:      autocert.DirCache("certs"),            //folder for storing certificates
				Email:      config.mail,
			}
			go http.ListenAndServe(":http", certManager.HTTPHandler(nil)) // 支持 http-01
			server := &http.Server{
				Addr: ":https",
				TLSConfig: &tls.Config{
					GetCertificate: certManager.GetCertificate,
					NextProtos:     []string{http2.NextProtoTLS, "http/1.1"},
					MinVersion:     tls.VersionTLS12,
				},
				MaxHeaderBytes: 32 << 20,
			}
			go server.ListenAndServeTLS("", "")
		} else {
			go http.ListenAndServe(fmt.Sprintf(":%d", config.prot), nil)
		}
		log.Println("Start ybot in webhook mode")
		startBot(updates, bot)
	} else {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates, err := bot.GetUpdatesChan(u)
		if err != nil {
			log.Printf("Init pull mode error %s", err.Error())
			return
		}
		log.Println("Start ybot in pull mode")
		//add google cloud run support
		if port := os.Getenv("PORT"); port != "" {
			// run a fake http server
			http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("hello ybot"))
			})
			go http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
		}
		startBot(updates, bot)
	}

}
