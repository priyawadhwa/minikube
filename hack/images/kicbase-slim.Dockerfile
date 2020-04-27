ARG COMMIT_SHA
# using base image created by kind https://github.com/kubernetes-sigs/kind/blob/master/images/base/Dockerfile
# which is an ubuntu 19.10 with an entry-point that helps running systemd
# could be changed to any debian that can run systemd
FROM kindest/base:v20200317-92225082 as base
USER root

# specify version of everything explicitly using 'apt-cache policy'
RUN apt-get update && apt-get install -y --no-install-recommends \
    lz4=1.9.1-1 \
    gnupg=2.2.12-1ubuntu3 \ 
    sudo=1.8.27-1ubuntu4.1 \
    docker.io=19.03.2-0ubuntu1 \
    openssh-server=1:8.0p1-6build1 \
    dnsutils=1:9.11.5.P4+dfsg-5.1ubuntu2.1 \
    && rm /etc/crictl.yaml \
    && rm -rf /var/lib/apt/lists/*

# disable non-docker runtimes by default
RUN systemctl disable containerd
# enable docker which is default
RUN systemctl enable docker
# making SSH work for docker container 
# based on https://github.com/rastasheep/ubuntu-sshd/blob/master/18.04/Dockerfile
RUN mkdir /var/run/sshd
RUN echo 'root:root' |chpasswd
RUN sed -ri 's/^#?PermitRootLogin\s+.*/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN sed -ri 's/UsePAM yes/#UsePAM yes/g' /etc/ssh/sshd_config
EXPOSE 22
# create docker user for minikube ssh. to match VM using "docker" as username
RUN adduser --ingroup docker --disabled-password --gecos '' docker 
RUN adduser docker sudo
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
USER docker
RUN mkdir /home/docker/.ssh
USER root
# kind base-image entry-point expects a "kind" folder for product_name,product_uuid
# https://github.com/kubernetes-sigs/kind/blob/master/images/base/files/usr/local/bin/entrypoint
RUN mkdir -p /kind
# Deleting leftovers
RUN apt-get clean -y && rm -rf \
  /var/cache/debconf/* \
  /var/lib/apt/lists/* \
  /var/log/* \
  /tmp/* \
  /var/tmp/* \
  /usr/share/doc/* \
  /usr/share/man/* \
  /usr/share/local/* \
  RUN echo "kic! Build: ${COMMIT_SHA} Time :$(date)" > "/kic.txt"
