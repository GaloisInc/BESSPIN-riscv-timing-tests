import numpy as np
from scipy.interpolate import Rbf
from scipy.interpolate import griddata
import matplotlib
matplotlib.use('Agg')

import matplotlib.pyplot as plt
from matplotlib import cm

records = []

base = 0
diff = 1 << 63

min_val = 0
max_val = min_val + diff

file_handle = open("records.out", "r")
for line in file_handle:
    if len(line) != 0:
        fields = line.split()

        left_operand = int(fields[0], 16)
        right_operand = int(fields[1], 16)
        cycle_count = float(fields[3])

        if left_operand >= min_val and left_operand < max_val and right_operand >= min_val and right_operand < max_val:
            records.append((left_operand, right_operand, cycle_count))

unique_records = list(set(records))

x_list = [ record[0] for record in unique_records ]
y_list = [ record[1] for record in unique_records ]
z_list = [ record[2] for record in unique_records ]

max_operand = max(x_list)
if max(x_list) < max(y_list):
    max_operand = max(y_list)

x = np.array(x_list)
y = np.array(y_list)
z = np.array(z_list)

ti = np.linspace(min_val, max_val, 256)
xi, yi = np.meshgrid(ti, ti)

# rbf = Rbf(x, y, z, epsilon=2)
# zi = rbf(xi, yi)

zi = griddata((x,y), z, (xi,yi), method = 'cubic')

plt.scatter(x, y, 10, z, cmap = cm.jet)
# plt.pcolor(xi, yi, zi, cmap = cm.jet)
# plt.title('RBF interpolation - multiquadrics')
plt.xlim(min_val, max_val)
plt.ylim(min_val, max_val)
plt.colorbar()
plt.savefig('rbf2d.png')
