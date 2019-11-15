
files=["vm", "minikube", "kubeadm"]
dirs=["1", "2", "3"]

for f in files:
    for d in dirs:
        with open(d+"/"+f+".log") as infile, open(d+"/"+f+".csv", 'w') as outfile:
            first = True
            for line in infile:
                if first:
                    first = False
                    continue
                outfile.write("\n")
                outfile.write(" ".join(line.split()).replace(' ', ','))
                outfile.write(",") # trailing comma shouldn't matter