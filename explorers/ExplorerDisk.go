package explorer

import (
	"github.com/shirou/gopsutil/disk"
	"os"
	"reflect"
	//"os"
	"time"

	logrusRotate "github.com/LazarenkoA/LogrusRotate"
	"github.com/prometheus/client_golang/prometheus"
)

type (
	ExplorerDisk struct {
		BaseExplorer
	}
)

func (this *ExplorerDisk) Construct(s Isettings, cerror chan error) *ExplorerDisk {
	logrusRotate.StandardLogger().WithField("Name", this.GetName()).Debug("Создание объекта")

	this.summary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: this.GetName(),
			Help: "Показатели дисков",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"host", "disk", "metrics"},
	)

	this.settings = s
	this.cerror = cerror
	prometheus.MustRegister(this.summary)
	return this
}

func (this *ExplorerDisk) StartExplore() {
	delay := reflect.ValueOf(this.settings.GetProperty(this.GetName(), "timerNotyfy", 10)).Int()
	logrusRotate.StandardLogger().WithField("delay", delay).WithField("Name", this.GetName()).Debug("Start")

	timerNotyfy := time.Second * time.Duration(delay)
	this.ticker = time.NewTicker(timerNotyfy)
	host, _ := os.Hostname()

FOR:
	for {
		this.Lock()
		func() {
			logrusRotate.StandardLogger().WithField("Name", this.GetName()).Trace("Старт итерации таймера")
			defer this.Unlock()

			dinfo, err := disk.IOCounters()
			if err != nil {
				logrusRotate.StandardLogger().WithField("Name", this.GetName()).WithError(err).Error()
				return
			}

			this.summary.Reset()
			for k, v := range dinfo {
				this.summary.WithLabelValues(host, k, "WeightedIO").Observe(float64(v.WeightedIO))
			}
		}()

		select {
		case <-this.ctx.Done():
			break FOR
		case <-this.ticker.C:
		}
	}
}

func (this *ExplorerDisk) GetName() string {
	return "disk"
}
