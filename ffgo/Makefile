all:
	gcc -c ffgocodec.c -o ./ffgocodec.o -I ./ffmpeg/include
	ar rvs ./ffgo.a ./ffgocodec.o
#    export GO111MODULE=off
	 go build libff.go
clean:
	rm -fr ./ffgocodec.o ./ffgo.a