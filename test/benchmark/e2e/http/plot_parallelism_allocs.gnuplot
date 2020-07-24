set datafile separator comma
set datafile missing NaN
set xlabel "Parallelism"
set ylabel "Memory Allocated/Ops"
payload_size_kb=ARG1
payload_size=payload_size_kb*1024
output_senders=(exist("ARG2") && ARG2 != ""?ARG2:1)

print "Plotting with payload size ".payload_size." and output_senders ".output_senders.""

plot "baseline-binary.csv" using 1:(($2==payload_size && $3==output_senders)?$5:1/0) title "Baseline Binary ".payload_size_kb."kb" with linespoint, \
     #"baseline-structured.csv" using 1:($2==payload_size && $3==output_senders?$5:1/0) title "Baseline Structured ".payload_size_kb."kb" with linespoint, \
     #"binding-structured-to-structured.csv" using 1:($2==payload_size && $3==output_senders?$5:1/0) title "Binding Structured to Structured ".payload_size_kb."kb" with linespoint, \
     #"binding-structured-to-binary.csv" using 1:($2==payload_size && $3==output_senders?$5:1/0) title "Binding Structured to Binary ".payload_size_kb."kb" with linespoint, \
     #"binding-binary-to-structured.csv" using 1:($2==payload_size && $3==output_senders?$5:1/0) title "Binding Binary to Structured ".payload_size_kb."kb" with linespoint, \
     #"binding-binary-to-binary.csv" using 1:($2==payload_size && $3==output_senders?$5:1/0) title "Binding Binary to Binary ".payload_size_kb."kb" with linespoint, \
     #"client-binary.csv" using 1:($2==payload_size && $3==output_senders?$5:1/0) title "Client Binary ".payload_size_kb."kb" with linespoint, \
     #"client-structured.csv" using 1:($2==payload_size && $3==output_senders?$5:1/0) title "Client Structured ".payload_size_kb."kb" with linespoint
pause -1
