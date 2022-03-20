#!/bin/bash
go mod tidy
go build main.go
echo "build success,you can run by following methods:"
echo "./main"
echo "cat input.txt | ./main"
