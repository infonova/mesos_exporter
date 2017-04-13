# Prometheus Mesos Exporter
Exporter for Mesos master and agent metrics for __Mesos > 1.0__


## Using
The Mesos Exporter can either expose cluster wide metrics from a master or task
metrics from an agent.

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
```

Usually you would run one exporter with `-master` pointing to the current
leader and one exporter for each slave with `-slave` pointing to it. You should 
be able to run the mesos-exporter like this:

- Master: `mesos-exporter -master http://mesos-master.local:5050`
- Agent: `mesos-exporter -slave http://mesos-slave.local:5051`
