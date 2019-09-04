package main

import (
	"flag"
)

func init() {
	flag.StringVar(&config.token, "ybot-token", "", "Telegram bot token")
	flag.StringVar(&config.domain, "ybot-domain", "", "Y2bot domain")
	flag.StringVar(&config.outputPath, "ybot-dir", "./", "Specify the output path")
	flag.IntVar(&config.threadNumber, "ybot-thread", 2, "The number of process thread")
	flag.IntVar(&config.prot, "ybot-port", 8008, "The port number of server")
	flag.BoolVar(&config.acme, "ybot-acme", true, "Print extracted data")
	flag.BoolVar(&config.hook, "ybot-hook", true, "webhook or pull")
	flag.StringVar(&config.mail, "ybot-mail", "", "Specify the output path")
}

func main() {
	parse()
	config.check()
	start()
}
