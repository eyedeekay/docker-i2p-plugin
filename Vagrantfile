# -*- mode: ruby -*-
# vi: set ft=ruby :


# THIS IS BROKEN AND I DON'T USE VAGRANT. It's going to work **last**.
# thanks for your patience.
# eyedeekay

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

$bootstrap=<<SCRIPT
apt-get update
apt-get -y install wget bridge-utils i2p i2p-router
wget -qO- https://experimental.docker.com/ | sh
service docker stop
gpasswd -a vagrant docker
ovs-vsctl add-br ovsbr-docker0
ovs-vsctl set-manager ptcp:6640
echo DOCKER_OPTS=\\"--default-network=ovs:ovsbr-docker0\\" >> /etc/default/docker
service docker restart
mkdir -p /usr/share/docker/plugins
touch /run/docker/plugins/ovs.sock
wget -O /home/vagrant/docker-i2p-plugin https://github.com/eyedeekay/docker-i2p-plugin/raw/master/binaries/docker-i2p-plugin-0.1-Linux-x86_64
chmod +x /home/vagrant/docker-i2p-plugin
SCRIPT

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

  config.vm.define "ovsnode" do |ovs|
    ovs.vm.box = "ubuntu/trusty64"
    ovs.vm.hostname = "ovsnode"
    ovs.vm.network :private_network, ip: "192.168.33.10"
    ovs.vm.provider "virtualbox" do |vb|
     vb.customize ["modifyvm", :id, "--memory", "1024"]
    end
    ovs.vm.provision :shell, inline: $bootstrap
  end

end
