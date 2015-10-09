#! /bin/bash

# defaults
file=$(pwd)/monitor.log
name="./app"
file_flag=false
name_flag=false


for i in "$@"
do

case $i in
    -f|--file)
      file_flag=true
    ;;
    -n|--name)
      name_flag=true
    ;;
    *)
      if $file_flag
      then
        file="$(pwd)/${i}"
        file_flag=false
      fi

      if $name_flag
      then
        name="${i}"
        name_flag=false
      fi
    ;;
esac
done

if [ ! -e "$file" ] ; then
    touch "$file"
fi

if [ ! -w "$file" ] ; then
    echo cannot write to $file
    exit 1
fi

echo 'pid	psr	c	pcpu	state	nlwp	vsz	rss	size	%mem	seconds' > $file

cleanup () {
  kill -s SIGTERM $!
  echo
  echo "exiting..."
  exit 0
}

trap cleanup SIGINT SIGTERM

while [ 1 ]
do
  ps -C $name -L -o pid,psr,c,pcpu,state,nlwp,vsz,rss,size,%mem | sed 1d | sed "s/ \+/\t/g" | sed "s/$/\t`date +%s`/" >> $file
  sleep 1 &
  wait $!
done
