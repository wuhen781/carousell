#!/bin/bash
go mod tidy
go build main.go
if [ $? -ne 0 ]
	then
echo " build fails, you can also directly run the compiled binary(main)" 
	else
echo "build success,you can run by following methods:"
echo "./main"
echo "cat input.txt | ./main"
 fi
