import csv
import matplotlib.pyplot as plt
import numpy as np

udpTime = []
quicTime = []
with open('../sendCycle.csv', newline='') as csvfile:
    rows = csv.reader(csvfile)
    i = 0
    for row in rows:
        if i > 0:
            udpTime.append(int(row[1]))
            quicTime.append(int(row[0]))
        i += 1

# draw UDP send cycle CDF
timeInterval = 500
timeEnd = 30000.0
x = np.arange(4000.0,timeEnd,timeInterval)

y = x*0
timeIndex = 0

timeEndInt = int(timeEnd)
i = 0
for value in udpTime:
    i+= 1
    if value >= timeEndInt:
        print("udp index: ", i)
        break
    if value >= x[timeIndex] and value < x[timeIndex] + timeInterval:
        y[timeIndex] += 1
        continue
    while value >= x[timeIndex] + timeInterval:
        timeIndex+=1

sum = 0
i = 0
for value in y:
    sum += value
    y[i] = sum
    i += 1


w = x*0
timeIndex = 0

i = 0
for value in quicTime:
    i += 1
    if value >= timeEndInt:
        print("quic index: ", i)
        break
    if value >= x[timeIndex] and value < x[timeIndex] + timeInterval:
        w[timeIndex] += 1
        continue
    while value >= x[timeIndex] + timeInterval:
        timeIndex+=1

sum = 0
i = 0

for value in w:
    sum += value
    w[i] = sum
    i += 1

y = y/5000
w = w/5000

# time: [4000,21000]
# x = [1,2,3,4,5]
# y = [5,10,20,35,40]
fig =plt.figure()
plt.plot(x,y,color = 'black',label = 'UDP')
plt.ylim(0,1)
plt.xlim(0,30100)
plt.plot(x,w,color = 'red',label = 'QUIC')
plt.title('3GPP downlink cycle time CDF')
plt.legend()
plt.xlabel('Nanoseconds')
plt.ylabel('Cumulative probability')

plt.savefig('p1.png')

