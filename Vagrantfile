# -*- mode: ruby -*-
# vi: set ft=ruby :

require 'yaml'

boxes = YAML.load_file(File.join(File.dirname(__FILE__), '.dev/boxes.yaml'))

Vagrant.configure("2") do |config|
  boxes.each do |box|
    config.vm.define "#{box.gsub("/", "-")}" do |node|
      node.vm.box = box
      node.vm.provider "virtualbox" do |v|
        v.memory = 256
      end
    end
  end
end
