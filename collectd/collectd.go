/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collectd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

/*
Start http server with info from first metrics config
handler transforms from incoming http request to metric to send
on a channel to streamIt()
*/

var (
	metricChan = make(chan []plugin.Metric)
)

type RelayCollector struct {
}

func (RelayCollector) StreamMetrics(mts []plugin.Metric) (<-chan []plugin.Metric, error) {
	ch := make(chan []plugin.Metric)
	go streamIt(mts, ch)
	port := 9999
	go runMyServer(port)
	return ch, nil
}

func runMyServer(port int) {
	http.HandleFunc("/metrics", requestToMetric)
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func requestToMetric(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "err: %v", err)
		return
	}

	// raw request
	fmt.Printf(string(body))
	// Example data format for each line:
	//PUTVAL wcp2/swap/swap-used interval=10.000 1484002777.039:0.000000
	rm := strings.Replace(string(body), "\r", "", -1)
	s := strings.Split(rm, "\n")
	var mts []plugin.Metric
	// foreach line, we want to parse it
	for _, line := range s {
		if len(line) == 0 {
			continue
		}
		// split on spaces so we can get each section individually
		sp := strings.Split(line, " ")
		// Skip PUTVAL
		//first section is hostname + namespace, separated by /
		path := strings.Split(sp[1], "/")
		namespace := path[1:]
		hostname := path[0]
		tags := make(map[string]string)
		tags["hostname"] = hostname
		// interval=interval
		// discard interval for now
		// data is format epoch-time:data
		dtime := strings.Split(sp[3], ":")
		data, err := strconv.ParseFloat(dtime[1], 64)
		if err != nil {
			fmt.Println("ERROR:", err, "\n\n\n")
			continue
		}
		// last section is time since epoch in seconds
		seconds := strings.Split(dtime[0], ".")
		t, err := strconv.ParseInt(seconds[0], 10, 64)
		if err != nil {
			fmt.Println("ERROR:", err, "\n\n\n")
			continue
		}
		timestamp := time.Unix(t, 0)

		metric := plugin.Metric{
			Namespace: plugin.NewNamespace(namespace...),
			Timestamp: timestamp,
			Version:   1,
			Data:      data,
			Tags:      tags,
		}
		fmt.Println("Adding metric: ", namespace)
		mts = append(mts, metric)
	}

	metricChan <- mts
}

func streamIt(mts []plugin.Metric, ch chan []plugin.Metric) {
	for metrics := range metricChan {
		ch <- metrics
	}
}

func (RelayCollector) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}

	metric := plugin.Metric{
		Namespace: plugin.NewNamespace("snap-relay", "collectd"),
		Version:   1,
	}
	metrics = append(metrics, metric)
	return metrics, nil
}

func (RelayCollector) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	return *policy, nil
}
