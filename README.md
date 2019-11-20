# ![logo](https://raw.githubusercontent.com/HybridDevTools/d3nver/master/docs/images/denver_logo_small.png) D3NVER

D3nver helper application

## Requirements

Until now, only VirtualBox 6 is supported as provider so please make sure to have VirtualBox 6+ running on your computer.

For Windows users, you must add VirtualBox to your PATH system environment variable, you can find the procedure here : https://www.build-business-websites.co.uk/add-vboxmanage-to-path

## Geting started

[Download](https://github.com/HybridDevTools/d3nver/releases) and uncompress the release anywhere on your computer then, open a terminal (you can use Powershell for Windows users) in this directory.

```bash
./denver
Denver, the Developer ENVironment

Usage:
  denver [command]

Available Commands:
  help        Help about any command
  init        Init the instance
  ssh         Connect through ssh in local terminal
  start       Start the instance
  status      Check if the instance is ready to operate
  stop        Stop the instance
  term        Connect through configured terminal

Flags:
      --config string   config file (default is ./conf/config.yml)
  -h, --help            help for denver
      --version         version for denver

Use "denver [command] --help" for more information about a command.
```

### Set your configuration file

Rename or copy the `config.yml.dist` to `config.yml` in the `conf` folder.
Take time to adjust your settings in the `conf/config.yml` file, especially :

- `channel` : either `stable` or `beta`. Only use `beta` if you are testing new Denver features.
- `instance` : adjust Denver's performances, especially memory and CPUs allocation (the more the better but beware to not eat too much resources from your workstation. Don't try to allocate more than half the memory/cpu cores you have on your workstation so for a 16GB/8 cores laptop, don't allocate more than 8GB/4 cpus to Denver)
- `userinfo` : your credentials and dont't forget your SSH key (used for cloning from our repositories)

### Initialize your instance

Once your configuration file is ready, get into the denver folder and open a terminal :

```bash
# Mac/Linux users
./denver init

# Windows users
denver.exe init
```

In this phase, the `denver` will :
* check if a Denver instance already exists and if not, download the DENVER's RBI from the selected channel (beta/stable)
* uncompress the RBI image locally
* create Denver's virtual machine on the selected provider (VirtualBox)

This operation has to be done just once.

### Start your instance

Once Denver instance has been initialized, you can start it :

```bash
# Mac/Linux users
./denver start

# Windows users
denver.exe start
```

### Connect to DENVER

#### Through SSH

There are many ways to connect to the DENVER instance through SSH.
First, if you provided the path to your own SSH key in the `config.yml` file, it had been pushed and authorized inside the instance so that you can connect from your terminal passwordless.

`denver` also provide 2 alternative ways to connect to the instance.

```bash
# connect with your own terminal and SSH key
ssh ldevuser@10.10.10.10

# ----------------------------------------------------------------------------

# connect with denver and the configured terminal from config.yml (Mac/Linux)
./denver term

# connect with denver and the configured terminal from config.yml (Windows)
denver.exe term

# ----------------------------------------------------------------------------

# connect with denver with the current terminal (Mac/Linux)
./denver ssh

# connect with denver with the current terminal (Windows)
denver.exe ssh

```

#### Shared volume

Working with Denver means storing your source files inside the Denver instance itself and not on your local workstation.
However, that doesn't mean that you can't use your preferred local IDE or editor.
Inside the Denver instance, there is a `Projects` folder, this is the place where your sources should be stored.
This folder is shared with your workstation through multiple protocols and depending your operating system, you can choose the best fit :

- NFS : recommended for Mac/Linux, this is by far the best option. To mount the share folder on your workstation, you must use this command line

  ```bash
  # Linux users
  sudo mkdir /media/$USER/denver
  sudo mount -t nfs 10.10.10.10:/home/ldevuser/Projects /media/$USER/denver

  # Mac users
  sudo mkdir /private/denver
  sudo mount -t nfs -o resvport,rw,noatime,rsize=32768,wsize=32768,timeo=10 10.10.10.10:/home/ldevuser/Projects /private/denver

  ```

- Samba : recommended for Windows

  ```bash
  # Windows users, use cmd or Powershell, not bash
  NET USE Z: \\10.10.10.10\Projects
  ```
