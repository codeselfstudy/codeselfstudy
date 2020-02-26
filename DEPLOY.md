# Deployment

This file contains notes on deployment that will be moved into the wiki docs after they are working.

## Provision Server

Create a DigitalOcean droplet with Ubuntu 18.04 as the distro.

SSH into the server:

```text
$ ssh root@<ip_address>
```

and update the packages (it's a root prompt):

```text
# apt update && apt upgrade
```


```text
# apt install make
```

Install Docker with the instructions found [here](https://docs.docker.com/install/linux/docker-ce/ubuntu/#install-docker-engine---community-1).

Install Docker Compose with [these instructions](https://docs.docker.com/compose/install/).

## Set Timezone

```text
# dpkg-reconfigure tzdata
```

## Create User


As root:

```text
# adduser <username>
# usermod -aG sudo <username>
```

Configure the new user to use Docker without `sudo`: [instructions](https://docs.docker.com/install/linux/linux-postinstall/#manage-docker-as-a-non-root-user).

## Change SSH Port


```text
# vim /etc/ssh/sshd_config
```

In that file, uncomment the `Port 22` line and change `22` to something else and set `PasswordAuthentication` to `no`.

Disable IPv6 for SSH and restart SSH:

```text
# echo 'AddressFamily inet' | sudo tee -a /etc/ssh/sshd_config
# sudo systemctl restart sshd
```

[NOTE: for some reason, changing the port to 4444 isn't working, and it's still running on Port 22, even after rebooting.]

Copy your SSH key onto the server for the new user:

```text
# ssh-copy-id <username>@<ip-address>
```

## Install fail2ban

```text
# apt-get install fail2ban
```

## Clone the Repo

```text
$ git clone https://github.com/codeselfstudy/codeselfstudy.git
```

If you need to use another branch:

```text
$ git checkout <branch_name>
```

## Install Node

Install [nvm](https://github.com/nvm-sh/nvm#installing-and-updating).

Then install node 12:

```text
$ nvm install 12
```
