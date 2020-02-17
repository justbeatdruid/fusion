package logs

import (
	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
)

type requestLogger struct {
	prefix string
	level  klog.Level
}

func (r *requestLogger) SetPrefix(prefix string) {
	r.prefix = prefix
}

func (r *requestLogger) Printf(format string, v ...interface{}) {
	klog.V(r.level).Infof(format, v...)
}

func (r *requestLogger) Println(v ...interface{}) {
	klog.V(r.level).Infof("%+v", v)
}

func GetGoRequestLogger(level int) gorequest.Logger {
	return &requestLogger{
		level: klog.Level(level),
	}
}
