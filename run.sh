#!/bin/bash

# this shell script executes several runs of the performance tests

numberOfClients=10
maxValueTemplate=255

for templateLength in {64..128..16}
do
	echo $templateLength
	go test -args $numberOfClients $templateLength $maxValueTemplate noipfeTiming
	sleep 60    #let the processor cool down a bit
done

sleep 300

for templateLength in {64..128..16}
do
	echo $templateLength
	go test -args $numberOfClients $templateLength $maxValueTemplate ipfeTiming
	sleep 60    #let the processor cool down a bit
done
