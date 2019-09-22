package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/backend"
	"github.com/docker/docker/api/types/plugins/logdriver"
	"github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/daemon/logger/gelf"
	"github.com/docker/docker/daemon/logger/jsonfilelog"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	protoio "github.com/gogo/protobuf/io"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/tonistiigi/fifo"
)

const (
	driverName = "gelf-multi"
)

var gelfOpt = []string{"gelf-address", "gelf-compression-type", "gelf-compression-level", "gelf-tcp-max-reconnect", "gelf-tcp-reconnect-delay", "tag", "labels", "labels-regex", "env", "env-regex"}
var jsonOpt = []string{"max-file", "max-size", "compress", "tag", "labels", "labels-regex", "env", "env-regex"}

type driver struct {
	mu     sync.Mutex
	logs   map[string]*dockerInput
	idx    map[string]*dockerInput
	logger log.Logger
}

type dockerInput struct {
	stream      io.ReadCloser
	gelfLoggers []logger.Logger
	jsonLogger  logger.Logger
	info        logger.Info
	logger      log.Logger
}

func (l *dockerInput) Close() {
	if err := l.stream.Close(); err != nil {
		level.Error(l.logger).Log("msg", "error while closing fifo stream", "err", err)
	}
	for _, gelfLogger := range l.gelfLoggers {
		if err := gelfLogger.Close(); err != nil {
			level.Error(l.logger).Log("msg", "error while closing gelf logger", "err", err)
		}
	}
	if err := l.jsonLogger.Close(); err != nil {
		level.Error(l.logger).Log("msg", "error while closing json logger", "err", err)
	}
}

func newDriver(logger log.Logger) *driver {
	return &driver{
		logs:   make(map[string]*dockerInput),
		idx:    make(map[string]*dockerInput),
		logger: logger,
	}
}

func newGelfLogger(logCtxOriginal *logger.Info, num string) (*logger.Info, error) {
	var logCtx logger.Info
	copier.Copy(&logCtx, &logCtxOriginal)
	logCtx.Config = make(map[string]string)
	for _, v := range gelfOpt {
		if val, ok := logCtxOriginal.Config["gelf-multi-"+v+"."+num]; ok {
			logCtx.Config[v] = val
		}
	}
	if err := gelf.ValidateLogOpt(logCtx.Config); err != nil {
		return &logCtx, err
	}
	return &logCtx, nil
}

func newJSONLogger(logCtxOriginal *logger.Info) (*logger.Info, error) {
	var logCtx logger.Info
	copier.Copy(&logCtx, &logCtxOriginal)
	logCtx.Config = make(map[string]string)
	for _, v := range jsonOpt {
		if val, ok := logCtxOriginal.Config["json-multi-"+v]; ok {
			logCtx.Config[v] = val
		}
	}
	if err := jsonfilelog.ValidateLogOpt(logCtx.Config); err != nil {
		return &logCtx, err
	}
	return &logCtx, nil
}

func (d *driver) StartLogging(file string, logCtx logger.Info) error {
	d.mu.Lock()
	if _, exists := d.logs[file]; exists {
		d.mu.Unlock()
		return fmt.Errorf("logger for %q already exists", file)
	}
	d.mu.Unlock()

	if logCtx.LogPath == "" {
		logCtx.LogPath = filepath.Join("/var/log/docker", logCtx.ContainerID)
	}

	if err := os.MkdirAll(filepath.Dir(logCtx.LogPath), 0755); err != nil {
		return errors.Wrapf(err, "error setting up logger dir\n")
	}
	var logJCtx, err = newJSONLogger(&logCtx)
	jsonLogger, err := jsonfilelog.New(*logJCtx)
	if err != nil {
		return errors.Wrap(err, "error creating jsonfile logger\n")
	}

	level.Debug(d.logger).Log("msg", "Start logging", "id", logCtx.ContainerID, "file", file, "logpath", logCtx.LogPath)

	var gelfLoggers []logger.Logger
	gelfCount, err := strconv.Atoi(logCtx.Config["gelf-count"])
	if err != nil {
		return errors.Wrap(err, "error creating gelf multi logger - gelf-count not valid\n")
	}
	for i := 0; i < gelfCount; i++ {
		var num = strconv.Itoa(i)
		var logGCtx, err = newGelfLogger(&logCtx, num)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error preparing gelf logger %d\n", i))
		}
		gelfLogger, err := gelf.New(*logGCtx)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error creating gelf logger %d\n", i))
		}
		gelfLoggers = append(gelfLoggers, gelfLogger)
	}

	f, err := fifo.OpenFifo(context.Background(), file, syscall.O_RDONLY, 0700)
	if err != nil {
		return errors.Wrapf(err, "error opening logger fifo: %q", file)
	}

	d.mu.Lock()
	lf := &dockerInput{f, gelfLoggers, jsonLogger, logCtx, d.logger}
	d.logs[file] = lf
	d.idx[logCtx.ContainerID] = lf
	d.mu.Unlock()
	go consumeLog(lf)
	return nil

}

func (d *driver) StopLogging(file string) error {
	level.Debug(d.logger).Log("msg", "Stop logging", "file", file)
	d.mu.Lock()
	lf, ok := d.logs[file]
	if ok {
		lf.Close()
		delete(d.logs, file)
	}
	d.mu.Unlock()
	return nil
}

func consumeLog(lf *dockerInput) {
	dec := protoio.NewUint32DelimitedReader(lf.stream, binary.BigEndian, 1e6)
	defer dec.Close()
	defer lf.Close()
	var buf logdriver.LogEntry
	for {
		if err := dec.ReadMsg(&buf); err != nil {
			if err == io.EOF || err == os.ErrClosed || strings.Contains(err.Error(), "file already closed") {
				level.Debug(lf.logger).Log("msg", "shutting down log logger", "id", lf.info.ContainerID, "err", err)
				return
			}
			dec = protoio.NewUint32DelimitedReader(lf.stream, binary.BigEndian, 1e6)
		}
		var msgOriginal, msgTemp logger.Message
		msgOriginal.Line = buf.Line
		msgOriginal.Source = buf.Source
		if buf.PartialLogMetadata != nil {
			if msgOriginal.PLogMetaData == nil {
				msgOriginal.PLogMetaData = &backend.PartialLogMetaData{}
			}
			msgOriginal.PLogMetaData.ID = buf.PartialLogMetadata.Id
			msgOriginal.PLogMetaData.Last = buf.PartialLogMetadata.Last
			msgOriginal.PLogMetaData.Ordinal = int(buf.PartialLogMetadata.Ordinal)
		}
		msgOriginal.Timestamp = time.Unix(0, buf.TimeNano)

		//Loops through all gelf loggers
		for _, gelfLogger := range lf.gelfLoggers {
			copier.Copy(&msgTemp, &msgOriginal)
			if err := gelfLogger.Log(&msgTemp); err != nil {
				level.Error(lf.logger).Log("msg", "error pushing message to gelf", "id", lf.info.ContainerID, "err", err, "message", msgTemp)
			}
		}
		copier.Copy(&msgTemp, &msgOriginal)
		if err := lf.jsonLogger.Log(&msgTemp); err != nil {
			level.Error(lf.logger).Log("msg", "error writing log message", "id", lf.info.ContainerID, "err", err, "message", msgTemp)
		}

		buf.Reset()
	}
}

func (d *driver) Name() string {
	return driverName
}

func (d *driver) ReadLogs(info logger.Info, config logger.ReadConfig) (io.ReadCloser, error) {
	d.mu.Lock()
	lf, exists := d.idx[info.ContainerID]
	d.mu.Unlock()
	if !exists {
		return nil, fmt.Errorf("logger does not exist for %s", info.ContainerID)
	}

	r, w := io.Pipe()
	lr, ok := lf.jsonLogger.(logger.LogReader)
	if !ok {
		return nil, fmt.Errorf("logger does not support reading")
	}

	go func() {
		watcher := lr.ReadLogs(config)

		enc := protoio.NewUint32DelimitedWriter(w, binary.BigEndian)
		defer enc.Close()
		defer watcher.ConsumerGone()

		var buf logdriver.LogEntry
		for {
			select {
			case msg, ok := <-watcher.Msg:
				if !ok {
					w.Close()
					return
				}

				buf.Line = msg.Line
				buf.Partial = msg.PLogMetaData != nil
				buf.TimeNano = msg.Timestamp.UnixNano()
				buf.Source = msg.Source

				if err := enc.WriteMsg(&buf); err != nil {
					_ = w.CloseWithError(err)
					return
				}
			case err := <-watcher.Err:
				_ = w.CloseWithError(err)
				return
			}

			buf.Reset()
		}
	}()

	return r, nil
}
