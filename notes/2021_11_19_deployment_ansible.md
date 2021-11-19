# Ansible deployment

## Local Testing

To test deployment locally, we use classic
[vagrant](https://www.vagrantup.com/). Use
[vagrant-vbguest](https://github.com/dotless-de/vagrant-vbguest) to keep guest
additions versions in sync.

```
$ vagrant version
Installed Version: 2.2.19
Latest Version: 2.2.19

You're running an up-to-date version of Vagrant!

$ vagrant plugin install vagrant-vbguest
$ vagrant init debian/buster64
```

Memory needs to be increased, compilation of prerequisites got killed with just
a 1024MB.

```ruby
  config.vm.provider "virtualbox" do |vb|
    # ...
    vb.memory = "4096"
  end
```

```
$ time vagrant up
```

