package flylog

import "go.uber.org/zap"

type LogProxy struct {
	Name string
}

func (lp *LogProxy) Write(p []byte) (n int, err error) {
	Logger.Info(lp.Name, zap.String("content", string(p)))
	return 0, nil
}

func NewLogProxy(name string) *LogProxy {
	return &LogProxy{
		Name: name,
	}
}
