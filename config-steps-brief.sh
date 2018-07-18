# brief summary of commands for my approach to running LivePeer transcoder node

# this is just an overview and this script is NOT meant to be executed,
# it will not work, and not all commands are meant to be run on the same instance.

# instance launch
aws --profile notation ec2 run-instances \
    --cli-input-json file://livepeer-transcoder-ec2-config.json  

# allocate elastic ip
aws --profile notation ec2 allocate-address  
aws --profile notation ec2 associate-address --instance-id <instance id> --public-ip <ip address> 

# get most recent LivePeer release
cd ~
curl -s -L https://github.com/livepeer/go-livepeer/releases/download/0.2.4/livepeer_linux.tar.gz > livepeer_linux.tgz
gzip -d -c livepeer_linux.tar.gz | tar xvf -

# prepare LivePeer volume
# if the LivePeeer EBS volume was not created at instance instantiation, create and attach now
aws --profile notation ec2 create-volume --size 100 --region us-east-1 --availability-zone us-east-1a --volume-type gp2  
aws --profile notation ec2 attach-volume --device /dev/sdg --instance-id <instance-id> --volume-id <volume-id>  

# login to the instance and create filesystem, mount point, and add volume to fstab (device names may vary):
sudo mkfs.ext4 /dev/xvdg  
sudo mkdir /d1  
echo "UUID=<volume UUID> /d1 ext4 defaults 0 2" | sudo tee -a /etc/fstab  
sudo mount /d1 

# prepare geth volume
# If the geth EBS volume was not created at instance instantiation, create and attach now.
aws --profile notation ec2 create-volume --size 500 --region us-east-1 --availability-zone us-east-1a --volume-type gp2  
aws --profile notation ec2 attach-volume --device /dev/sdh --instance-id <instance-id> --volume-id <volume-id>  

# login to the instance and create filesystem, mount point, and add volume to fstab (device names may vary):
sudo mkfs.ext4 /dev/xvdh  
sudo mkdir /d2     
echo "UUID=<volume UUID> /d2 ext4 defaults 0 2" | sudo tee -a /etc/fstab  
sudo mount /dev/xvdh /d2   

# set hostname
sudo hostname tc001.mydomain.com
# add fqdn to /etc/hosts
# and replace contents of /etc/hostname (with only hostname, not FQDN)

# setup directories
sudo mkdir -p /d1/livepeer/logs  
sudo mv -i ~/livepeer_linux /d1/livepeer/bin  
sudo chown -R ubuntu:ubuntu /d1/livepeer  
cd /d1

# check out repo
git clone git@github.com:alexlines/livepeer-transcoder-ops.git

# raise open filehandle limits
echo "ubuntu soft nofile 50000" | sudo tee -a /etc/security/limits.conf
echo "ubuntu hard nofile 50000" | sudo tee -a /etc/security/limits.conf

# edit /etc/pam.d/login and add or uncomment the line:
session required /lib/security/pam_limits.so

# install geth
sudo apt-get install -y software-properties-common
sudo add-apt-repository -y ppa:ethereum/ethereum
sudo apt-get update
sudo apt-get install -y ethereum 

# setup geth directories and copy config files into place
sudo mkdir /d2/geth-data
sudo chown -R ubuntu:ubuntu /d2/geth-data
sudo cp /d1/livepeer-transcoder-ops/private/config/geth/systemd/geth.service /etc/systemd/system/
sudo cp /d1/livepeer-transcoder-ops/private/config/geth/geth-config.toml /d2/geth-data/

# copy any existing .ethereum files or keys into place now in /d2/geth-data/.ethereum

# enable geth under systemd and start geth
sudo systemctl enable geth
sudo systemctl start

# check the status and logs
sudo systemctl status geth
sudo journalctl -u geth.service -f


# If you are going to use existing LivePeer account data, go ahead and copy it into place now in /d1/livepeer/.lpData/

# copy systemd unit file for LivePeer into place
sudo cp /d1/livepeer-transcoder-ops/private/config/livepeer/systemd/livepeer-transcoder.service /etc/systemd/system/
sudo systemctl enable livepeer-transcoder

# start LivePeer using systemd
sudo systemctl start livepeer-transcoder

# check status and watch the logs
sudo systemctl status livepeer-transcoder
sudo journalctl -u livepeer-transcoder.service -f

# now use the livepeer command line utility to enroll as a transcoder and set transcoder config:
# Choose 13. Invoke multi-step "become a transcoder"

/d1/livepeer/bin/livepeer_cli


