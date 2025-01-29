# cpu-mem-scheduler

## scheduler logical

- scheduler actively watch pod status and select the node for pod
- watch pod status, found pending pod and scheduler match this scheduler
- get node cpu and mem metrics, and sum it value, select the little one
- pod bind to node

## simple scheduler relay on cpu and memory

This is a simple scheduler relay on cpu and memory.

## Run

```
go run main.go
```
