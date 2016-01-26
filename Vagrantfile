# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
    config.vm.box = "modernIE/w10-edge"
    config.vm.provision "shell", path: "bootstrap.cmd"
    config.vm.synced_folder ".", "/gopath/src/github.com/lvillani/just-install"
    config.vm.synced_folder "~/.ssh", "/Users/vagrant/Desktop/ssh"

    config.vm.provider "virtualbox" do |v|
      v.gui = true
    end
end
