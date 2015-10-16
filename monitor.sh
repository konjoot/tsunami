#! /bin/bash

# defaults
file=$(pwd)/monitor.log
name="./app"

# helper flags
file_flag=false
name_flag=false

# looping over argument list
for i in "$@"
do

case $i in
    # if -f or --file arg met
    -f|--file)
      # then set file_flag
      file_flag=true
    ;;
    # if -n or --name arg met
    -n|--name)
      # then set name_flag
      name_flag=true
    ;;
    # in other cases
    *)
      # if file_flag is set
      if $file_flag
      then
        # set file from current arg
        file="$(pwd)/${i}"
        # and unset file_flag
        file_flag=false
      fi

      # if name_flag is set
      if $name_flag
      then
        # set file from current arg
        name="${i}"
        # and unset name_flag
        name_flag=false
      fi
    ;;
esac
done

# create file if not present
if [ ! -e "$file" ] ; then
    touch "$file"
fi

# check if file is writable
if [ ! -w "$file" ] ; then
    echo cannot write to $file
    exit 1
fi

# write head to the file
echo 'pid	user	pr	ni	virt	res	shr	s	pcpu	pmem	time	command	tcpu(us)	tcpu(sy)	tmem(total)	tmem(used)	seconds' > $file

# on exit callback
cleanup () {
  kill -s SIGTERM $!
  echo
  echo "exiting..."
  exit 0
}

# catching SIGINT SIGTERM
trap cleanup SIGINT SIGTERM

# run top in batch mode for pids which corresponds to name, format it's output with sed and write to file
top -p $(pidof $name | sed 's/[ tab]\+/,/g' -u) -b -d1 | sed '
/^top.*/ {d}       #drop each line which starts with "top"
/^Tasks.*/ {d}     #drop each line which starts with "Tasks"
/^Swap.*/ {d}      #drop each line which starts with "Swap"
/^[ tab]*PID/ {d}  #drop each line which starts with "PID"
/^[ tab]*$/ {d}    #drop each blank line

# from each line which starts with "Cpu" get cpu usage for user space(us) and for system (sy) and remember them
/^Cpu.*/ {s/.*Cpu(s):[ tab]\+\([0-9]\{1,3\}\.[0-9]\{1,2\}\)%us,.*\([0-9]\{1,3\}\.[0-9]\{1,2\}\)%sy.*/\1 \2/;x;d}

# from each line which starts with "Mem" get total memory count and used count, remember them, calculate current unix-timestamp and remember it too
/^Mem.*/ {s/.*Mem:[ tab]\+\([0-9]\+[a-zA-Z]*\)[ tab]\+total,[ tab]\+\([0-9]\+[a-zA-Z]*\)[ tab]\+used,.*/\1 \2/;H;s/\($\)/\t/;s/\(.*\)/date +%s/e;H;d}

# for each line which starts with digits add stored strings, remove all new line except last one, and switch all groups of spaces to tabs
/^[ tab]*[0-9]/ {s/^[ tab]\+//;G;s/\([ tab]\+\n\|\n\)/ /g;s/ \+/\t/g}' -u >> $file
