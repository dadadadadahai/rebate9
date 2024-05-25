package main

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

// 日志库
type LogStruct struct {
	file *os.File
	log  *logrus.Logger
	day  int
}

func (this *LogStruct) Init() {
	this.log = logrus.New()
	this.day = 0
	this.FileCreate()
	this.log.SetFormatter(this)
}
func (this *LogStruct) SetLevel(level logrus.Level) {
	this.log.SetLevel(level)
}

func (this *LogStruct) Info(msg string) {
	this.FileCreate()
	this.log.Info(msg)
}
func (this *LogStruct) Warning(msg string) {
	this.FileCreate()
	this.log.Warning(msg)
}
func (this *LogStruct) Error(msg string) {
	this.FileCreate()
	this.log.Error(msg)
}
func (this *LogStruct) Debug(msg string) {
	this.FileCreate()
	this.log.Error(msg)
}

// 根据时间判断是否产生新的日志
func (this *LogStruct) FileCreate() {
	day := time.Now().Day()
	if this.day != day {
		//执行重新初始化
		filename := time.Now().Format("20060102")
		filename = filename + ".log"
		this.file.Close()
		this.file, _ = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		writes := []io.Writer{
			this.file,
			os.Stdout,
		}
		fileAndStdoutWriter := io.MultiWriter(writes...)
		this.log.SetOutput(fileAndStdoutWriter)
		this.day = day
	}
}
func (this *LogStruct) Format(entry *logrus.Entry) ([]byte, error) {
	timename := time.Now().Format("2006-01-02 15:04:05")
	return []byte(entry.Level.String() + " " + timename + " " + entry.Message + "\r\n"), nil
}
