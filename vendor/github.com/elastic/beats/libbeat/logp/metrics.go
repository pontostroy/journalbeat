package logp

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/elastic/beats/libbeat/monitoring"
)

var Globalstr = "Preparing"

// logMetrics logs at Info level the integer expvars that have changed in the
// last interval. For each expvar, the delta from the beginning of the interval
// is logged.
func logMetrics(metricsCfg *LoggingMetricsConfig) {
	if metricsCfg.Enabled != nil && *metricsCfg.Enabled == false {
		Info("Metrics logging disabled")
		return
	}
	if metricsCfg.Period == nil {
		metricsCfg.Period = &defaultMetricsPeriod
	}
	Info("Metrics logging every %s", metricsCfg.Period)

	ticker := time.NewTicker(*metricsCfg.Period)

	prevVals := monitoring.MakeFlatSnapshot()
	for range ticker.C {
		snapshot := snapshotMetrics()
		delta := snapshotDelta(prevVals, snapshot)
		prevVals = snapshot

		if len(delta) == 0 {
			Info("No non-zero metrics in the last %s", metricsCfg.Period)
			Globalstr = noMetrics()
			continue

		}

		metrics := formatMetrics(delta)
		Info("Non-zero metrics in the last %s:%s", metricsCfg.Period, metrics)
		Globalstr = formatMetricsHttp(delta)
	}
}

// LogTotalExpvars logs all registered expvar metrics.
func LogTotalExpvars(cfg *Logging) {
	if cfg.Metrics.Enabled != nil && *cfg.Metrics.Enabled == false {
		return
	}

	zero := monitoring.MakeFlatSnapshot()
	metrics := formatMetrics(snapshotDelta(zero, snapshotMetrics()))
	Info("Total non-zero values: %s", metrics)
	Info("Uptime: %s", time.Now().Sub(startTime))
	Globalstr = formatMetricsHttp(snapshotDelta(zero, snapshotMetrics()))
}

func snapshotMetrics() monitoring.FlatSnapshot {
	return monitoring.CollectFlatSnapshot(monitoring.Default, monitoring.Full, true)
}

func snapshotDelta(prev, cur monitoring.FlatSnapshot) map[string]interface{} {
	out := map[string]interface{}{}

	for k, b := range cur.Bools {
		if p, ok := prev.Bools[k]; !ok || p != b {
			out[k] = b
		}
	}

	for k, s := range cur.Strings {
		if p, ok := prev.Strings[k]; !ok || p != s {
			out[k] = s
		}
	}

	for k, i := range cur.Ints {
		if p := prev.Ints[k]; p != i {
			out[k] = i - p
		}
	}

	for k, f := range cur.Floats {
		if p := prev.Floats[k]; p != f {
			out[k] = f - p
		}
	}

	return out
}

func formatMetrics(ms map[string]interface{}) string {
	keys := make([]string, 0, len(ms))
	for key := range ms {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	var buf bytes.Buffer
	for _, key := range keys {
		buf.WriteByte(' ')
		buf.WriteString(key)
		buf.WriteString("=")
		buf.WriteString(fmt.Sprintf("%v", ms[key]))
	}
	return buf.String()
}
func formatMetricsHttp(ms map[string]interface{}) string {
	keys := make([]string, 0, len(ms))
	for key := range ms {
		keys = append(keys, key)
	}

	mkeys := make(map[string]interface{})
	mkeys["libbeat.logstash.publish.read_bytes"] = "0"
	mkeys["libbeat.logstash.publish.write_bytes"] = "0"
	mkeys["libbeat.logstash.call_count.PublishEvents"] = "0"
	mkeys["libbeat.logstash.published_and_acked_events"] = "0"
	mkeys["libbeat.publisher.messages_in_worker_queues"] = "0"
	mkeys["libbeat.publisher.published_events"] = "0"

	pkeys := make([]string, 0, len(mkeys))
	for pkey := range mkeys {
		pkeys = append(pkeys, pkey)
	}

	//sort.Strings(keys)
	var buf bytes.Buffer
	for _, key := range keys {
		switch key {
		case "libbeat.logstash.publish.read_bytes":
			mkeys[key] = ms[key]
		case "libbeat.logstash.publish.write_bytes":
			mkeys[key] = ms[key]
		case "libbeat.logstash.call_count.PublishEvents":
			mkeys[key] = ms[key]
		case "libbeat.logstash.published_and_acked_events":
			mkeys[key] = ms[key]
		case "libbeat.publisher.messages_in_worker_queues":
			mkeys[key] = ms[key]
		case "libbeat.publisher.published_events":
			mkeys[key] = ms[key]
		}
	}
	sort.Strings(pkeys)
	for _, pkey := range pkeys {
		buf.WriteString(pkey)
		buf.WriteString(":")
		buf.WriteByte(' ')
		buf.WriteString(fmt.Sprintf("%v", mkeys[pkey]))
		buf.WriteByte('\n')
	}

	return buf.String()
}

func noMetrics() string {
	var buf bytes.Buffer
	buf.WriteString("libbeat.logstash.call_count.PublishEvents: 0\n")
	buf.WriteString("libbeat.logstash.publish.read_bytes: 0\n")
	buf.WriteString("libbeat.logstash.publish.write_bytes: 0\n")
	buf.WriteString("libbeat.logstash.published_and_acked_events: 0\n")
	buf.WriteString("libbeat.publisher.messages_in_worker_queues: 0\n")
	buf.WriteString("libbeat.publisher.published_events: 0\n")
	return buf.String()
}
