**Running a transcoder**  
  * Master reference docs https://livepeer.readthedocs.io/en/latest/getting_started.html  
  * Also this thread on the forum, [Transcoder Megathread - Start here to learn about playing the role of transcoder on Livepeer](https://forum.livepeer.org/t/transcoder-megathread-start-here-to-learn-about-playing-the-role-of-transcoder-on-livepeer/190)  
  * Also [this transcoder bash setup script](https://gist.github.com/ChrisChiasson/206b2500d1792135ef7e41dc825f8122), posted to discord by Chris Chiasson  
  * For GPU capabilities, consider [P2 GPU instances](https://aws.amazon.com/ec2/instance-types/p2/) (crazy expensive) and [Elastic GPUs](https://aws.amazon.com/ec2/elastic-gpus/details/) which can be attached to certain instance types.   
  * This will spin up a [c4.2xlarge](https://www.ec2instances.info/?filter=c4.2xlarge&cost_duration=monthly) instance in us-east with 15GB RAM, 8vCPUs, "High" network perf, [EBS optimized](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSOptimized.html), and a 32GB [gp2 standard SSD](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumeTypes.html#EBSVolumeTypes_gp2) running an [EBS-backed Ubuntu 18.04 LTS image](https://cloud-images.ubuntu.com/locator/ec2/). Cost ~$300/month (on-demand).  
    * Considering increasing storage size, 100GB would be ~$10/month and 500GB would be ~$50. 100GB is probably fine as long as you don't run a geth node locally.  
    * Should probably attach a dedicated EBS volume for livepeer anyway ...  
  * I'm using the [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/installing.html) to launch instances with [this configuration](https://gist.github.com/alexlines/f8a83c4705755b74e7592e686a4832e9)  
  * **Note** This command line won't work for you as-is because the named profile "notation" won't exist on your system. You need to [create your own named profile config](https://docs.aws.amazon.com/cli/latest/userguide/cli-multiple-profiles.html) and reference that. This config also references named security groups which you won't have (which just allow ssh from certain sources) and has "DryRun" set to true (change to false to actually launch an instance), so adjust accordingly.  


```
aws --profile notation ec2 run-instances \
    --cli-input-json file://livepeer-transcoder-ec2-config.json
```  


**Ops TODO**  
  * Dedicated EBS Volume  
  * Set up a [public elastic ip](https://docs.aws.amazon.com/cli/latest/reference/ec2/run-instances.html) and document and operationalize  
  * Raise filehandle limit  https://forum.livepeer.org/t/increase-file-limit-as-a-transcoder/170 and [see my own notes](https://gist.github.com/alexlines/dc870ce77cbd754ee6aca67898cafa10) and document and operationalize        
    * Process supervisor to keep livepeer running (or restart periodically) - systemd, etc.    
  * Make sure timesync is active. In 18.04, base system uses [systemd-timesyncd](https://www.freedesktop.org/software/systemd/man/timedatectl.html) which may be fine, but probably want to use chrony for better accuracy and syncing.  
   * Some [Chrony](https://chrony.tuxfamily.org/) links: [FAQ](https://chrony.tuxfamily.org/faq.html), also [Time Sync](https://help.ubuntu.com/lts/serverguide/NTP.html) in Ubuntu, and a [basic chrony config overview](https://blog.ubuntu.com/2018/04/09/ubuntu-bionic-using-chrony-to-configure-ntp).  
  * Run LP on a dedicated attached EBS vol, not the default root vol ...  
  * What livepeer / ipfs / etc logs needs to be rotated?   
  * Oh, so livepeerjs is the main official API (via RPC)? Kindof missed that https://github.com/livepeer/livepeerjs/tree/master/packages/sdk ... oh, so it's for interacting with LP smart contracts, sot it doesn't expose reward(), but maybe could use to monitor whether transcoder x has successfully called reward() this round? I wonder if you could also just call it via mycrypto? doesn't look like reward is exposed via the ABI that doug sent us. Can definitely use it to get info about a specific transcoder with `getTranscoder('0xf00...')` can also use it to set transcoder parameters, which is useful, but strange you can't call reward?  
  * Making sure that `reward()` gets called is a priority  
  * Maybe just use ELB's for health checks (not sure about classic ELB vs ALB yet)  
    * https://www.sumologic.com/aws/elb/aws-elastic-load-balancers-classic-vs-application/  
    * https://docs.aws.amazon.com/elasticloadbalancing/latest/application/introduction.html   
  * [Dockerize?](https://github.com/livepeer/docker-livepeer)  
  * What about using [vault](https://www.hashicorp.com/blog/using-vault-to-build-an-ethereum-wallet) or something for private keys?  
  * DNS name?  


**LivePeer questions**  
* The most complicated part is knowing the correct steps and order to take in the CLI to make sure the transcoder is active and how to debug if it isn't, also what options to start it with. There isn't an official walkthrough of recommended arguments to start LP with and then register on mainnet as a transcoder.  
* Sane recommended -gasLimit to start with? -> sounds like omitting gasPrice flag might be the way to go after 0.2.3, which will rely on the gas oracle instead.  
* Setting up an automatic call to `reward()` once per round. It looks like it calls it automatically at the beginning of every round, but it can fail - if connection to Ethereum node isn't good, if gas prices are high, etc. If it doesn't get called, the newly minted LPT are lost, so it's v important to check.  
  * Can it be called via the http interface?  
  * Is it ok to call it more than once per round?  
* Worth setting up a dedicated ipfs node in local network?  
* Is it worth it to run with GPU? How much does it help? What specifically leverages the GPU - ffmpeg? 
* Best way to backup the account / credentials tied to the node?  
* What livepeer / ipfs / etc logs needs to be rotated?  
* Nice to have: Make sure `initializeRound()` has been called - cannot call `reward()` until it has  
* What ports should be open? Open to the world?  
  * Video Ingest Endpoint - rtmp://localhost:1935  
  * livepeer_cli params: --http value local http port (default: "8935")  - this is a control port via http, I would make sure this is protected, can set configs here, bond(), etc.  
* yep, this is the port to poll to track status and info. Available commands are documented in [webserver.go](https://github.com/livepeer/go-livepeer/blob/ec288f43b60fbf3bd61f81b636538b5b004aaa86/server/webserver.go)  
* Seems like it can send server metrics to http://viz.livepeer.org:8081/metrics ? see [livepeer.go](https://github.com/livepeer/go-livepeer/blob/master/cmd/livepeer/livepeer.go) interesting that it can end metrics by default, wonder if that can be redirected and to what kind of server. you can also specify the monitor host to send to  
  * Maybe to the monitor server here? https://github.com/livepeer/go-livepeer/blob/master/monitor/monitor.go  
  * Looks like there's a separate monitor server project https://github.com/livepeer/streamingviz  ... although it hasn't been touched in a year  
  * Or publish them to CloudWatch? https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/publishingMetrics.html  
    * https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/customize-containers-cw.html  
* livepeer_cli params: --rtmp value local rtmp port (default: "1935")  
  * Testnet vs mainnet  
* Some of these could go into FAQ  
* What's the .eth name service to translate from name -> eth address?  
* How to import existing ETH account / keys? Maybe a [current bug](https://github.com/livepeer/go-livepeer/issues/304)  
* Is there anything to backup?  
* Guidelines on setting up basic monitoring / alerting  
* Custom nagios or cloudwatch plugin (possible?) to do health check requests (maybe ELB?) and maybe check basic stats  
  * Go and systemd both support watchdog http://0pointer.de/blog/projects/watchdog.html  
  * Any admin interface available via network - http, etc? Or do you need to build an http request -> livepeer_cli  
* Securing your node and access to private ETH key  
* Running a local geth node would definitely be helpful
  * Although it's not 100% clear in the docs - offical docs [recommend running geth](https://livepeer.readthedocs.io/en/latest/node.html) but other info, such as [this forum post](https://forum.livepeer.org/t/how-to-run-livepeer-with-geth/143), say it's not necessary. I know that it's not required but I think it's clearly beneficial, eg:  
  * "If your connection to the Ethereum network is not always great, causing instability in your transcoder and leading to errors such as "Error with x:EOF" so it's better to run your own geth / parity node - ideally not on the same box either. You can use --ethIpcPath flag to specify the local IPC file location, which is a much more stable way to connect to the Ethereum network."  
  * How to specify a local geth/parity node that's on the same network but maybe not the same box? ok looks like you can also specify ethUrl := flag.String("ethUrl", "", "geth/parity rpc or websocket url") from [livepeer.go](https://github.com/livepeer/go-livepeer/blob/master/cmd/livepeer/livepeer.go#L83)  
  * You can specify a local geth node on the command line via `-ethDatadir` flag when starting the node. The directory specified should contain the ipc file for the Geth node, from https://github.com/livepeer/wiki/wiki/Livepeer-Node   
  * See this post for running a local geth instance https://forum.livepeer.org/t/transcoder-tip-geth-light-client/247/7  
  * Need a full copy of ETH blockchain? It seems a fast sync is sufficient  
  * My preference is to run it on a dedicated local node (not the transcoder)  
* Unclear from docs: needs ffmpeg? the specially built static version? https://github.com/livepeer/ffmpeg-static  
* What do you need to do to transfer your transcoder identity to a new box? eg if you need to migrate hardware for some reason?  
* How can you run multiple transcoder instances, behind a load balancer, for example, but have them all use the same identity? Because you just register as a single transcoder id, right?  


**Becoming an active transcoder on mainnet**  
* **spin up a fresh node but try to put an old account in place before starting the livepeer binary**  
* Fund your node with ETH and LPT and bond to yourself  
* Specifying the Ethereum account - Eth Account Each Livepeer node should have an Ethereum account. Use -ethAccountAddr to specify the account address. You should make sure the keys to the account is in the keystore directory of ethDatadir you passed in.  


**Grab LivePeer binaries**  
  * You can build from scratch if you want but why ...
  * Download the latest mainnet-targeted livepeer and livepeer_cli from https://github.com/livepeer/go-livepeer/releases.  
```
curl -s -L https://github.com/livepeer/go-livepeer/releases/download/0.2.3/livepeer_linux.tgz > livepeer_linux.tgz
gzip -d -c livepeer_linux.tgz | tar xvf -
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


**Test ETH on Rinkeby**  
  * Figuring out the correct address to request test ETH to from Rinkeby faucet:
  * When starting `./livepeer --rinkeby` the output says:  
```
***Livepeer is running on the Rinkeby test network: 0x37dc71366ec655093b9930bc816e16e6b587f968***
``` 
  * but livepeer_cli in 'node status' shows:
```
*--------------*--------------------------------------------*
|  ETH Account | 0x0C47D7852c14001b78c157Fd6Fc8938488CD45F0 |
*--------------*--------------------------------------------*
```
  * and keystore filenames are: 
```
$ ls -1 ~/.lpData/keystore/
UTC--2018-05-02T19-12-28.040032202Z--0c47d7852c14001b78c157fd6fc8938488cd45f0
```
  * so to confirm, when I requested ETH from [rinkeby faucet](https://faucet.rinkeby.io/), I requested it be sent to address 0x0C47D7852c14001b78c157Fd6Fc8938488CD45F0 in this [g+ post](https://plus.google.com/+alexlines/posts/HesTiinUH9v) and it worked fine.  
  * To get test livepeer tokens (LPT), while `livepeer --rinkeby` is running in another terminal or under `screen`, run `livepeer_cli` and select option `10. Get test LPT` sometimes it fails and may need to request again, can watch the console log of `livpeer --rinkeby` to see status.  
  * To run as a transcoder. First kill the running process `livepeer --rinkeby` and then start it with transcoder flags:  
  ```
  ./livepeer --rinkeby --transcoder --publicIP <public ip>  
  ```
  * Then run the `./livepeer_cli` in another terminal to register as a transcoder
    * Choose `15. Become a transcoder`  and choose `PricePerSegment,` `FeeShare,` `BlockRewardCut,` and how much LPT to bond to yourself.  
  * Your transcoder will not become active until the next round starts - that's currently a period of 1 day.  
  * Check your status on the explore page https://explorer.livepeer.org/accounts/0x0C47D7852c14001b78c157Fd6Fc8938488CD45F0/transcoding  
  
**HTTP Query interface**  
  * `curl http://localhost:8935/getAvailableTranscodingOptions`  
  * Can also set broadcast config by POST'ing params to http://localhost:8935/setBroadcastConfig  
  
