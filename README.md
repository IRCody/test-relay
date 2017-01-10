# test-relay

POC for streaming + remote plugins using collectd.

# Demo

### node1
* collectd
* relay plugin

### node2
* Snapteld
* Mock file publisher plugin

Data is collected with collectd on node1, written to the relay plugin and then sent to Snapteld on node2 where it is written to a file by the publisher plugin.


# Setup

Use [remotey](https://github.com/IRCody/snap/tree/remotey) branch of my snap fork and [remotey](https://github.com/ircody/snap-plugin-lib-go/tree/remotey) branch of the plugin-lib-go fork.
These branches add a POC for support of Plugin Initiated Workflows + service(remote) plugins.

# Steps to run:
Install collectd (apt-get install collectd)

Modify the config file
`vim /etc/collectd/collectd.conf`

Make sure that `LoadPlugin write_http` is present with the following config:
```
<Plugin write_http>                                                                                                                                                                                                
        <URL "http://<hostname>:9999/metrics">                                                                                                                                                                                                                                                                                                                                                   
        </URL>                                                                                                                                                                                                     
</Plugin>        
```

Enable the other collector plugins in the conf file with what you want to see.



Build snapteld and this plugin. To build this plugin use:
`go build -o relay main.go` from the root of the repo.

Then: `./relay --service` to start as a service on port 8183.

Now start snapteld:
`./snapteld -t 0 -l 1`

Load Plugins:

```
snaptel plugin load "http://<hostname>:8183"
snaptel plugin load snap-plugin-publisher-mock-file
```
-- The example task uses the mock file publisher

Load the task, currently snaptel does extra validation around schedule types so load task via curl:

`curl -vX POST localhost:8181/v1/tasks -d @collectd.json --header "Content-Type: application/json"`


Watch the results:
`tail -f /tmp/collectd.log`


