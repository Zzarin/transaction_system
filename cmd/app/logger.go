package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func InitializeLoger() *zap.Logger {
	config := zap.NewProductionEncoderConfig()     //creates a new config for the zap encoder.
	config.EncodeTime = zapcore.ISO8601TimeEncoder //sets the time encoder to follow the standard ISO time format.
	// or 			zapcore.NewConsoleEncoder / zapcore.NewJSONEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)                                           //creates a file encoder that follows the standard JSON encoding.
	consoleEncoder := zapcore.NewConsoleEncoder(config)                                     //new config to write in the console
	logFile, _ := os.OpenFile("./logs/text.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666) //opens up the log.json in the Create and Append mode.
	writer := zapcore.AddSync(logFile)                                                      //creates a file writer that will be used later on by zap to write the messages to file.
	defaultLogLevel := zapcore.DebugLevel                                                   // create a variable that says the default logging level will be at Debug. You can change it.
	//add the file writer, encoder, and default log level to the zap core instance.
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel), //enable console
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.DebugLevel)) //creates a new Zap instance passing the previously created configs. It also adds in other
	// options like including the log caller, and the stack trace ( for the errors ).
	return logger
}
