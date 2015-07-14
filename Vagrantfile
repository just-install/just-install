# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
    config.vm.box = "lvillani/win2008r2"
    config.vm.provision "shell", path: "script/bootstrap.cmd"
    config.vm.synced_folder ".", "/gopath/src/github.com/lvillani/just-install"
    config.vm.synced_folder "~/.ssh", "/Users/vagrant/Desktop/ssh"
end
