# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
    config.vm.box = "modernIE/w7-ie8"
    config.vm.provision "shell", path: "bootstrap.cmd"
    config.vm.synced_folder ".", "/gopath/src/github.com/just-install/just-install"
    config.vm.synced_folder "~/.ssh", "/Users/vagrant/Desktop/ssh"

    config.vm.provider "virtualbox" do |v|
        v.gui = true
    end
end
