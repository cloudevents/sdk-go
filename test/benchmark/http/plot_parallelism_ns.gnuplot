set datafile separator comma
set datafile missing NaN
set xlabel "Parallelism"
set ylabel "Nanoseconds/Ops"
payload_size=ARG1
plot "baseline.csv" using 1:($2==payload_size?$3:1/0) title "Baseline ".payload_size."kb" with linespoint, "receiver-sender.csv" using 1:($2==payload_size?$3:1/0) title "Receiver Sender ".payload_size."kb" with linespoint, "pipe.csv" using 1:($2==payload_size?$3:1/0) title "Pipe ".payload_size."kb" with linespoint, "client.csv" using 1:($2==payload_size?$3:1/0) title "Client ".payload_size."kb" with linespoint
pause -1
