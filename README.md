# prometheus-exporter

[![Go Report Card](https://goreportcard.com/badge/github.com/prezhdarov/prometheus-exporter)](https://goreportcard.com/report/github.com/prezhdarov/prometheus-exporter)

Set of tools and libraries to help build prometheus exporters of diuerse kind..

In short this is set of functions I borrowed from [node_exporter](https://github.com/prometheus/node_exporter) and reworked so a basic API consuming structure can be sneaked into every collector created. At least that was the idea. Also borrowed the config from [Victoria Metrics](https://github.com/VictoriaMetrics/VictoriaMetrics) envflag. 

## How to use

The files in **example** folder are all that is necessary to build a working exporter.

### Main program (example-exporter.go)

This is the where main() lives. Follow the comments in the example-exporter.go to build an exporter. The command-line flags are completely optional, but handy. I borrowed most of these from node_exporter too..

### I call it the API (in api/example.go)

This is where the target consumption takes place. The client (APIClient) requires three methods - a login, a get and a logout functions - to authenticate, read and clean any loose ends for every scrape. Again follow the comments in file to create your own.

### A set of collectors (example has only one in collcectors/example.go)

This is a set of metrics collectors sharing a package name. Each collector executes concurrently and must have unique name and needs an update method which will be called by the Collect function.
