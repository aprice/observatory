#!/bin/bash
expvarmon -ports="13100" -vars="duration:Uptime,mem:memstats.Alloc,mem:memstats.Sys,mem:memstats.HeapInuse,duration:memstats.PauseNs,duration:memstats.PauseTotalNs,Goroutines"
