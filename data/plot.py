from matplotlib import pyplot as plt
import sys
import numpy as np

def plot():
    #This is the number of test-runs that are contained in the input file.
    runsPerFile = 2
    #This is the number of lines per run, i.e.
    linesPerRun = 5
    filename = 'runtime.txt'
    
    if len(sys.argv) == 4:
        runsPerFile = int(sys.argv[1])
        linesPerRun = int(sys.argv[2])
        filename = sys.argv[3]

    templateLength = [0] * linesPerRun
    clientRegistrationTime = [0] * linesPerRun
    serverRegistrationTime = [0] * linesPerRun
    clientAuthenticationTime = [0] * linesPerRun
    serverAuthenticationTime = [0] * linesPerRun
    ipfeClientRegistrationTime = [0] * linesPerRun
    ipfeClientAuthenticationTime = [0] * linesPerRun
    ipfeServerAuthenticationTime = [0] * linesPerRun
    
    file = open(filename, 'r')

    Lines = file.readlines()[1:]

    assert len(Lines) == runsPerFile * linesPerRun

    
    for runIndex in range(0,runsPerFile):
        for lineIndex in range(0, linesPerRun):
            line = Lines[runIndex * linesPerRun + lineIndex].strip()

            #process lines of our scheme
            splitStrings = line.split(';')
            templateLength[lineIndex] += (int(splitStrings[0]))
            
            if (int(splitStrings[1])) != -1:
                clientRegistrationTime[lineIndex] += (int(splitStrings[1]))
            if (int(splitStrings[2])) != -1:
                serverRegistrationTime[lineIndex] += (int(splitStrings[2]))
            if (int(splitStrings[3])) != -1:
                clientAuthenticationTime[lineIndex] += (int(splitStrings[3]))
            if (int(splitStrings[4])) != -1:
                serverAuthenticationTime[lineIndex] += (int(splitStrings[4]))
            if (int(splitStrings[5])) != -1:
                ipfeClientRegistrationTime[lineIndex] += (int(splitStrings[5]))
            if (int(splitStrings[6])) != -1:
                ipfeClientAuthenticationTime[lineIndex] += (int(splitStrings[6]))
            if (int(splitStrings[7])) != -1:
                ipfeServerAuthenticationTime[lineIndex] += (int(splitStrings[7]))
            
            
            
    for i in range(0, linesPerRun):
        templateLength[i] = int(templateLength[i] / runsPerFile)
        # the division by 2 is because half of the runs are the ipfe-Timings and the other half are the protocol timings
        clientRegistrationTime[i] /= (runsPerFile /2)
        serverRegistrationTime[i] /= (runsPerFile /2)
        clientAuthenticationTime[i] /= (runsPerFile /2)
        serverAuthenticationTime[i] /= (runsPerFile /2)
        ipfeClientRegistrationTime[i] /= (runsPerFile /2)
        ipfeClientAuthenticationTime[i] /= (runsPerFile /2)
        ipfeServerAuthenticationTime[i] /= (runsPerFile /2)


    #transform runtime from nanoseconds to milliseconds/seconds
    for i in range(0,len(clientRegistrationTime)):
        clientRegistrationTime[i] /= 1000000000
        serverRegistrationTime[i] /= 1000000000
        clientAuthenticationTime[i] /= 1000000000
        serverAuthenticationTime[i] /= 1000000000
        ipfeClientRegistrationTime[i] /= 1000000000
        ipfeClientAuthenticationTime[i] /= 1000000000
        ipfeServerAuthenticationTime[i] /= 1000000000
    
    max_list = [max(clientRegistrationTime), max(serverRegistrationTime), max(clientAuthenticationTime), max(serverAuthenticationTime)]
    max_value = max(max_list)
    
    plt.ylim(0,max_value*1.1)
    plt.plot(templateLength, clientRegistrationTime, 'r-')
    plt.plot(templateLength, clientRegistrationTime, 'ro', label='Enrollment (Client)')
    #plt.plot(templateLength, ipfeClientRegistrationTime, 'r--')
    #plt.plot(templateLength, ipfeClientRegistrationTime, 'ro', label='ipfe Enrollment (Client)')
    
    plt.plot(templateLength, serverRegistrationTime, 'g-')
    plt.plot(templateLength, serverRegistrationTime, 'gs', label='Enrolment (Server)')
    
    plt.plot(templateLength, clientAuthenticationTime, 'b-')
    plt.plot(templateLength, clientAuthenticationTime, 'bd', label='Authentication (Client)')
    #plt.plot(templateLength, ipfeClientAuthenticationTime, 'b--')
    #plt.plot(templateLength, ipfeClientAuthenticationTime, 'bd', label='ipfe Authentication (Client)')
    
    plt.plot(templateLength, serverAuthenticationTime, 'c-')
    plt.plot(templateLength, serverAuthenticationTime, 'c^', label='Authentication (Server)')
    #plt.plot(templateLength, ipfeServerAuthenticationTime, 'c--')
    #plt.plot(templateLength, ipfeServerAuthenticationTime, 'c^', label='ipfe Authentication (Server)')
    
    plt.xticks(ticks=templateLength)
    
    plt.legend(loc=0)
    plt.xlabel("Template length")
    plt.ylabel("Running time in seconds")
    plt.savefig("enrol-auth.svg")
    plt.clf()
    
    print(templateLength)
    
    #output values for the table in the paper
    print("Running time in seconds")
    print("Template length, Enrollment client (IPFE time), Enrollment server (IPFE time), Authentication client (IPFE time), Authentication server (IPFE time)")
    print(str(templateLength[0]) + ", "
        + str(clientRegistrationTime[0]) + " (" + str(ipfeClientRegistrationTime[0]) + "), "
        + str(serverRegistrationTime[0]) + " (0), "
        + str(clientAuthenticationTime[0]) + " (" + str(ipfeClientAuthenticationTime[0]) + "), "
        + str(serverAuthenticationTime[0]) + " (" + str(ipfeServerAuthenticationTime[0]) + ")")
    print(str(templateLength[4]) + ", "
        + str(clientRegistrationTime[4]) + " (" + str(ipfeClientRegistrationTime[4]) + "), "
        + str(serverRegistrationTime[4]) + " (0), "
        + str(clientAuthenticationTime[4]) + " (" + str(ipfeClientAuthenticationTime[4]) + "), "
        + str(serverAuthenticationTime[4]) + " (" + str(ipfeServerAuthenticationTime[4]) + ")")
    
    
    
    
plot()
