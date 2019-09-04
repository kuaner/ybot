package main

import (
	"flag"
)

func init() {
	flag.StringVar(&config.token, "ybot-token", "", "Telegram bot token")
	flag.StringVar(&config.domain, "ybot-domain", "", "Telegram webhook domain")
	flag.StringVar(&config.outputPath, "ybot-dir", "./", "Specify the output path")
	flag.IntVar(&config.threadNumber, "ybot-thread", 2, "The number of process thread")
	flag.IntVar(&config.prot, "ybot-port", 8008, "The port number of server")
	flag.BoolVar(&config.acme, "ybot-acme", true, "Enable autocert")
	flag.StringVar(&config.mail, "ybot-mail", "", "Specify acme mail address")
	flag.BoolVar(&config.hook, "ybot-hook", true, "Enable Webhook mode")
}

func main() {
	parse()
	config.check()
	start()
}
