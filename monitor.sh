#! /bin/bash

file=$(pwd)/${1-monitor.log}

if [ ! -e "$file" ] ; then
    touch "$file"
fi

if [ ! -w "$file" ] ; then
    echo cannot write to $file
    exit 1
fi

echo '' > $file

cleanup () {
  kill -s SIGTERM $!
  echo
  echo "exiting..."
  exit 0
}

trap cleanup SIGINT SIGTERM

while [ 1 ]
do
    ps -C kami -L -o pid,psr,c,pcpu,state,nlwp,vsz,rss,size,%mem,args | sed 1d | sed "s/$/ `date +%s`/" >> $file
    sleep 1 &
    wait $!
done
