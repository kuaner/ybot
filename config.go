package main

var config = cnf{}

type cnf struct {
	mail         string
	token        string
	domain       string
	outputPath   string
	hook         bool
	acme         bool
	hasFfmpeg    bool
	hasFfprobe   bool
	threadNumber int
	prot         int
}

func (c *cnf) check() error {
	c.hasFfmpeg = ffmpegExist()
	c.hasFfprobe = ffprobeExist()
	return nil
}
