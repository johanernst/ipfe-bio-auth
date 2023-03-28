# ipfe-bio-auth
This repository contains the source code of the paper "A Framework for UC Secure Privacy Preserving Biometric Authentication using Efficient Functional Encryption" (<--link coming soon-->). It is a two-factor authentication protocol. The first factor is the posession of a secret key stored on a secure hardware token. The second factor is the user's biometric. The protocol uses function-hiding inner-product functional encryption to let the server compute the distance of the biometric templates, while keeping the templates hidden from the server. 

## Running the code
To re-run the performance tests, execute ``./run.sh``
The measurements will be stored in data/runtime.txt

## Creating the plot
To create the plot of the running times  with the data from the paper, go to the folder "data" and execute ``python3 plot.py 6 5 runtime_paper.txt``
The first argument to plot.py is the number test-runs per file (running ``./run.sh`` creates two runs). The second argument is the number of measurements per run.
To create the plot after running the tests as described above, execute ``python3 plot.py``
The file enrol-auth.svg is the same as the plot in the paper. It has been created by running ``python3 plot.py 6 5 runtime_paper.txt``. The values in the table (from the paper) are the (commandline) output of running the aforementioned command.
