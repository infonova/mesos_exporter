# Prometheus Mesos Exporter
Prometheus Exporter for Mesos master and agent metrics. Requires __Mesos > 1.0__

## Using
The Mesos Exporter can either expose cluster wide metrics from a master or task
metrics from an agent. Usually you would run one exporter with `-master` pointing to the 
current leader and one exporter for each slave with `-slave` pointing to it. 

```sh
Usage of mesos-exporter:
  -addr string
       	Address to listen on (default ":9110")
  -ignoreCompletedFrameworkTasks
       	Don't export task_state_time metric
  -master string
       	Expose metrics from master running on this URL
  -slave string
       	Expose metrics from slave running on this URL
  -timeout duration
       	Master polling timeout (default 5s)
  -exportedTaskLabels
        Comma-separated list of task labels to include in the task_labels metric
  -trustedRedirects
        Comma-separated list of trusted hosts (ip addresses, host names) where metrics requests can be redirected
```

## Docker 
If you use docker, start the container like this (copy and paste code)
```
docker run  infonova/prometheus_mesos_exporter:1.0 -master http://mesos-master.local:5050
```
