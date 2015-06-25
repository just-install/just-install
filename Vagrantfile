# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
    config.vm.box = "lvillani/win2008r2"

    config.vm.provider "virtualbox" do |v|
        v.gui = true
        v.memory = 2048
    end

    config.vm.provider "vmware_fusion" do |v|
        v.gui = true
        v.vmx["memsize"] = "2048"
    end
end
