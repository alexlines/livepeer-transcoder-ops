**Running a transcoder**  
  * Master reference docs https://livepeer.readthedocs.io/en/latest/getting_started.html  
  * Also this thread on the forum, [Transcoder Megathread - Start here to learn about playing the role of transcoder on Livepeer](https://forum.livepeer.org/t/transcoder-megathread-start-here-to-learn-about-playing-the-role-of-transcoder-on-livepeer/190)  
  * Also [this transcoder bash setup script](https://gist.github.com/ChrisChiasson/206b2500d1792135ef7e41dc825f8122), posted to discord by Chris Chiasson  
  * For GPU capabilities, consider [P2 GPU instances](https://aws.amazon.com/ec2/instance-types/p2/) (crazy expensive) and [Elastic GPUs](https://aws.amazon.com/ec2/elastic-gpus/details/) which can be attached to certain instance types.    
  * This will spin up a [c4.2xlarge](https://www.ec2instances.info/?filter=c4.2xlarge&cost_duration=monthly) instance in us-east with 15GB RAM, 8vCPUs, "High" network perf, [EBS optimized](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSOptimized.html), and a 32GB [gp2 standard SSD](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumeTypes.html#EBSVolumeTypes_gp2). Cost ~$300/month (on-demand).  
  * I'm using the [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/installing.html) to launch instances with [this configuration](https://gist.github.com/alexlines/f8a83c4705755b74e7592e686a4832e9)  
  * **Note** This command line won't work for you as-is because the named profile "notation" won't exist on your system. You need to [create your own named profile config](https://docs.aws.amazon.com/cli/latest/userguide/cli-multiple-profiles.html) and reference that. This config also references named security groups which you won't have, so adjust accordingly.  


```
aws --profile notation ec2 run-instances \
    --cli-input-json file://livepeer-transcoder-ec2-config.json
```  


**Questions**  
  * Should I set up a public elastic ip?  
  * DNS name?  
  * What ports should be open? Open to the world?  
  * Raise filehandle limit  
  * Testnet vs mainnet  
  * What's the .eth name service to translate from name -> eth address?  
  * How to import an existing ETH account / keys? Might be a [current bug](https://github.com/livepeer/go-livepeer/issues/304)  
  * Guidelines on setting up basic monitoring / alerting  
    * Custom nagios or cloudwatch plugin (possible?) to do health check requests (maybe ELB?) and maybe check basic stats  
    * Is there an admin interface available via network - http, etc? Or do you need to build an http request -> livepeer_cli  
  * Securing your node and access to private ETH key  
  * Process supervisor to keep livepeer running (or restart periodically) - systemd, etc.  
  * Unclear from docs: need to run a local geth or not? https://forum.livepeer.org/t/how-to-run-livepeer-with-geth/143  
  * Unclear from docs: need to install ffmpeg? the specially built static version? https://github.com/livepeer/ffmpeg-static  


**ffmpeg**  
  * This section may be obsolete  
  * needs a specially-built (I think) statically compiled version of ffmpeg  
  * from this repo https://github.com/livepeer/ffmpeg-static  
  * can grab the linux x64 binary from this url and move it into $PATH  
```
cd
curl -s -L https://github.com/livepeer/ffmpeg-static/raw/master/bin/linux/x64/ffmpeg > ffmpeg
chmod 0755 ffmpeg
sudo chown root:root ffmpeg
sudo mv -i ffmpeg /usr/local/bin/
```

**Grab LivePeer binaries**  
  * You can build from scratch if you want but why ...
  * Download the latest mainnet-targeted livepeer and livepeer_cli from https://github.com/livepeer/go-livepeer/releases.  
```
curl -s -L https://github.com/livepeer/go-livepeer/releases/download/0.2.0/livepeer_linux.tar > livepeer_linux.tar
tar xvfp livepeer_linux.tar
cd livepeer_linux/
./livepeer
```


**Build from scratch if you must**  
  * See https://github.com/livepeer/go-livepeer  
```
sudo apt-get update
sudo apt-get install golang


# edit .profile:

  # edit the version number here as appropriate:
  export GOROOT="/usr/lib/go-1.6"
  export GOPATH="$HOME/goprojects"

  # set PATH so it includes user's private bin directories
  PATH="$HOME/bin:$HOME/.local/bin:$PATH:$GOPATH/bin"
  
mkdir goprojects

# ugh this isn't building right now, 
# just download livepeer_linux and livepeer_linux_cli from https://github.com/livepeer/go-livepeer/releases 
git clone git@github.com:livepeer/go-livepeer.git
cd go-livepeer
go get ./...
go build ./cmd/livepeer/livepeer.go

sudo apt-get install ffmpeg

# install local ethereum node, https://github.com/ethereum/go-ethereum/wiki/Building-Ethereum
sudo apt-get install software-properties-common
sudo add-apt-repository -y ppa:ethereum/ethereum
sudo apt-get update
sudo apt-get install ethereum
geth account new
# Address: {eee665a9f5bcb3a3e57c66571bbf144ab308d1cf} and pw: emigre smuggle lumbago
mkdir ~/.lpGeth
wget http://eth-testnet.livepeer.org/lptestnet.json
# if it complains about "field Genesis.number" delete the "number" field from lptestnet.json
geth --datadir ~/.lpGeth init lptestnet.json
# run geth in a screen session
screen
geth --datadir ~/.lpGeth --networkid 858585 --bootnodes "enode://080ebca2373d15762c29ca8d85ddc848f10a7ffc745f7110cacba4694728325d645292cb512d7168323bd0af1650fca825ff54c8dba20aec8878498fae3ff3c6@18.221.67.74:30303"
# exit the screen session
# run livepeer node in another screen session
screen
chmod 0755 livepeer_linux 
./livepeer_linux -testnet
# exit the screen session
./livepeer_cli_linux
```


**Add'l notes**  
  * get ether from crypto faucet in dashboard if necessary http://eth-testnet.livepeer.org
  * copy address of ether wallet display in livepeer_cli_linux at "Account Eth Addr"
  * paste it into a secret github gist, and copy the url of that gist into the faucet above
  * and select to get new ether
  * back in the livepeer_cli_linux, run "1." to get node status to see updated Eth balance
  * now get some test LivePeer tokens through the CLI, option "11"
  * check the token balance using the CLI
  * update with notes inside project dir  
  
  
