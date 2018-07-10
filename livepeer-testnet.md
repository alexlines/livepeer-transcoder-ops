## Running a LivePeer transcoder   
* Goals  
* Future Architecture Directions (see OPs TODO)  
* Config is code  
* Decisions  
platform - AWS, Linux, Ubuntu  
instance size  
why an elastic ip address  
- better flexibility
EBS Volumes - config and data concentration, separate from root disk, flexilibity - easily expandable, easily transferred to new instance (speed of recovery), easy to backup (EBS snapshots, easily automated)  
process supervision - systemd  
timekeeping - obviously important. In Ubuntu 18.04, the base system uses systemd-timesyncd which may be fine, but may want to consider using chrony for better accuracy and syncing **(grab links from below).**  
running a local geth node  
Security
* Automating startup of `livepeer` by automatically supplying a password for the ethereum?  
  * I just supplied a blank password the first time and then it doesn't ask for password on startup in the future  
  * What are implications of no password? For backing up files, for security of account in general, etc?  
  * Could just do it via command-line, but don't really want it to be visible to 'ps'  
  * Would prefer at least a config file if nothing else ... 
* Ports - all locked down, closed to the world and the local network except 4433, as required by [upcoming network updates](https://forum.livepeer.org/t/upcoming-networking-upgrades/298) which I think has to be open to the world.
  * Note that ssh is also closed to the world and, in our setup, only accessible through an ssh bastion host.  
* No root logins, auth via ssh keys only, keep security patches up-to-date, review all running procs and open ports, shut down (permanently) all unnecessary ones, make sure you have 2FA enable for your AWS account, backup regularly and automatically, monitor your boxes, regularly audit and rotate authorized ssh keys, AWS IAM permissions, sudo access, don't allow access via root AWS ssh keys, encourage audit trails via access via username-accounts (vs system accounts such as "ubuntu"), etc.  
  * [Publish metrics to CloudWatch](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/publishingMetrics.html)?  
    * Maybe with [custom events](https://aws.amazon.com/blogs/security/how-to-use-amazon-cloudwatch-events-to-monitor-application-health/)?  
    * https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/customize-containers-cw.html  
    * Monitor and document that reward() is called daily, publish to public cloudwatch dashboard? How to monitor?  
    * Easiest approach is by querying the local livepeer node via http but safest to make sure tx actually happened is to check the Ethereum blockchain
    * Can check if reward() was called by watching the latest transactions from the transcoder's account, either via the [etherscan API](https://etherscan.io/apis)
    * Can do the same thing by querying the local geth node -
      * Via geth console - but need to do this via API, maybe web3
      This tx was a call to reward()
      ```
      geth attach /d2/geth-data/.ethereum/geth.ipc
      > eth.getTransaction("0xcde8ec889fa7ed433d2a55c5f34f1be98f4dad97791a27c258d18eb1bad17d0f")
      > eth.getTransactionReceipt("0xcde8ec889fa7ed433d2a55c5f34f1be98f4dad97791a27c258d18eb1bad17d0f")
      ```
      * or via geth's [web3 javascript api](https://github.com/ethereum/wiki/wiki/JavaScript-API#web3ethgettransaction) for communicating with an ethereum node from inside a javascript application  
      * Or use the [JSON-RPC api](https://github.com/ethereum/wiki/wiki/JSON-RPC) directly. These docs include [curl examples](https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_gettransactionbyhash):
      ```
      curl -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionByHash","params":["0xcde8ec889fa7ed433d2a55c5f34f1be98f4dad97791a27c258d18eb1bad17d0f"],"id":1}' http://localhost:8545  
      ```
      but there's not an easy way to list recent transactions for an account or contract ... looks like filters/logs are the way to do this? https://github.com/ethereum/go-ethereum/issues/1897  or here https://github.com/ethereum/go-ethereum/issues/2104  
      * But would still need to decode the `input` param and translate it to the function name. According to the [Ethereum Contract ABI](https://github.com/ethereum/wiki/wiki/Ethereum-Contract-ABI), "the first four bytes of the call data for a function call specifies the function to be called. It is the first (left, high-order in big-endian) four bytes of the Keccak (SHA-3) hash of the signature of the function."   
        * GitHub: [ConsenSys ABI decoder](https://github.com/ConsenSys/abi-decoder) project    
        * GitHub: [Ethereum tx input data decoder project](https://github.com/miguelmota/ethereum-input-data-decoder)  
        * GitHub: [Decoder and encoder for the Ethereum ABI](https://github.com/ethereumjs/ethereumjs-abi)  
        * GitHub: [python eth abi](https://github.com/ethereum/eth-abi) project with input decoding [example on stackexchange](https://ethereum.stackexchange.com/questions/6297/python-eth-getfilterchanges-data-how-to-decode)  
      * The LivePeer process also makes an http api available, so it's possible to ask it for transcoder stats:  
      ```
      $ curl http://127.0.0.1:8935/transcoderInfo
      {"Address":"0x50d69f8253685999b4c74a67ccb3d240e2a56ed6","LastRewardRound":1018,"RewardCut":30000,"FeeShare":300000,"PricePerSegment":150000000000,"PendingRewardCut":30000,"PendingFeeShare":300000,"PendingPricePerSegment":150000000000,"DelegatedStake":6454553077282307328907,"Active":true,"Status":"Registered"}
      ```  
      such as `LastRewardRound` - the last round reward was called in. See [go-livepeer/server/webserver.go](https://github.com/livepeer/go-livepeer/blob/4589a1364fa9d29e9d196d259f1f235116d45953/server/webserver.go) for other functions you can call.  
      * Instead of decoding the contract input data could just string match since we know that the reward hex is "input":"0x228cb733", but it's a dirty hack.  
* Custom nagios or cloudwatch plugin (possible?) to do health check requests (maybe ELB?) and maybe check basic stats  
  * Go and systemd both support watchdog http://0pointer.de/blog/projects/watchdog.html  
* Securing your node and access to private ETH key  
  * Tradeoffs, hacks, etc.  
  * Storing private key in [AWS Parameter Store](https://aws.amazon.com/systems-manager/features/#Parameter_Store) in AWS [Key Management Service](https://aws.amazon.com/kms/)?  
  * [Getting started with AWS Parameter store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html) see also How AWS [Systems Manager Parameter Store Uses AWS KMS](https://docs.aws.amazon.com/kms/latest/developerguide/services-parameter-store.html?shortFooter=true)  
  * goog: [golang example aws kms](https://www.google.com/search?biw=1295&bih=1103&ei=N30yW4aHJ8vxzgLp1J-gBw&q=golang+example+aws+kms&oq=golang+example+aws+kms&gs_l=psy-ab.3..33i22i29i30k1.231231.237437.0.238301.22.17.0.4.4.0.305.1360.3j7j0j1.11.0....0...1c.1.64.psy-ab..8.14.1249...0j0i67k1j0i131i67k1j0i131k1j0i22i30k1.0.7Ap2nvkZiVw)  
  * It's just that the ethereum client doesn't seem to have the capability to get the account key over the network or from anything other than a file https://github.com/livepeer/go-livepeer/blob/master/eth/accountmanager.go  
* Running a local geth node would definitely be helpful
  * Although it's not 100% clear in the docs - official docs [recommend running geth](https://livepeer.readthedocs.io/en/latest/node.html) but other info, such as [this forum post](https://forum.livepeer.org/t/how-to-run-livepeer-with-geth/143), say it's not necessary. I know that it's not required but I think it's clearly beneficial, eg:  
  * "If your connection to the Ethereum network is not always great, causing instability in your transcoder and leading to errors such as "Error with x:EOF" so it's better to run your own geth / parity node - ideally not on the same box either. You can use --ethIpcPath flag to specify the local IPC file location, which is a much more stable way to connect to the Ethereum network."  
  * How to specify a local geth/parity node that's on the same network but maybe not the same box? ok looks like you can also specify ethUrl := flag.String("ethUrl", "", "geth/parity rpc or websocket url") from [livepeer.go](https://github.com/livepeer/go-livepeer/blob/master/cmd/livepeer/livepeer.go#L83)  
  * You can specify a local geth node on the command line via `-ethDatadir` flag when starting the node. The directory specified should contain the ipc file for the Geth node, from https://github.com/livepeer/wiki/wiki/Livepeer-Node   
  * See this post for running a local geth instance https://forum.livepeer.org/t/transcoder-tip-geth-light-client/247/7  
  * Need a full copy of ETH blockchain? It seems a fast sync is sufficient  
  * My preference is to run it on a dedicated local node (not the transcoder)  
  * Is it really ok to run geth light client vs fast-sync (or full node)?  


The goal is to run robust infrastructure for the LivePeer transcoding network. I care about  
* Availability, Performance, Security, Repeatability, Fast recovery   
* This is all very specific to AWS and Ubuntu. I haven't done the work to generalize for Amazon Linux, RHEL, CentOS, whatever.   

**Instance type and resources**  
I want to be sure this transcoder can perform, so for the initial phase I've overprovisioned the resources of CPU, RAM, disk performance, and bandwidth (details below). This means this specific configuration is expensive so feel free to choose lower-resource instance types.  

The instructions below will spin up an instance with the following characteristics:  

| | |  
| --- | --- |  
| Instance type | [c4.2xlarge](https://www.ec2instances.info/?filter=c4.2xlarge&cost_duration=monthly)  |  
| CPU | 8 vCPUs | 
| Network | High |
| EBS Optimized | [YES](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSOptimized.html) |
| OS | ami-85f9b8fa [Ubuntu 18.04 LTS HVM AMI](https://cloud-images.ubuntu.com/locator/ec2/) |
| Root disk | EBS-backed, 32GB [gp2 SSD](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumeTypes.html#EBSVolumeTypes_gp2) |
| EBS Vol 1 | 100GB gp2 SSD for LivePeer data |
| EBS Vol 2 | 500 GB gp2 SSD for dedicated local geth node |  

You can use the [AWS Command Line Interface](https://docs.aws.amazon.com/cli/latest/userguide/installing.html) to launch instances with these characteristics using [this configuration file](https://gist.github.com/alexlines/f8a83c4705755b74e7592e686a4832e9) as follows:  
**Note** This command line won't work for you as-is because the named profile "notation" won't exist on your system. You can [create your own named profile config](https://docs.aws.amazon.com/cli/latest/userguide/cli-multiple-profiles.html) and reference that. This config also references named security groups which you won't have (which just allow ssh from certain sources), a private key of a different name, and has "DryRun" set to true (change to false to actually launch an instance), so adjust accordingly.   

Launch the instance  
```
aws --profile notation ec2 run-instances \
    --cli-input-json file://livepeer-transcoder-ec2-config.json  
```  


Allocate elastic ip  to have a stable public address  
```
aws --profile notation ec2 allocate-address  
aws --profile notation ec2 associate-address --instance-id <instance id> --public-ip <ip address>  
```   


Grab LivePeer binaries  
You can build from scratch if you want but why ... I won't go into that, read more about it in the [official README](https://github.com/livepeer/go-livepeer/blob/master/README.md)    
Download the latest mainnet-targeted livepeer and livepeer_cli from https://github.com/livepeer/go-livepeer/releases.  
```
curl -s -L https://github.com/livepeer/go-livepeer/releases/download/0.2.4/livepeer_linux.tar.gz > livepeer_linux.tgz
gzip -d -c livepeer_linux.tar.gz | tar xvf -
cd livepeer_linux/
./livepeer
```


System Ops  
If the EBS volumes weren't created at instance instatiation, create them now.   
100GB gp2 disk for LivePeer storage / operations   
Adjust the availability zone to match instance az   
```
aws --profile notation ec2 create-volume --size 100 --region us-east-1 --availability-zone us-east-1a --volume-type gp2  
aws --profile notation ec2 attach-volume --device /dev/sdg --instance-id <instance-id> --volume-id <volume-id>  
# run locally on the box:  
sudo mkfs.ext4 /dev/xvdg  
sudo mkdir /d1  
echo "UUID=<volume UUID> /d1 ext4 defaults 0 2" | sudo tee -a /etc/fstab  
sudo mount /d1    
```  

500GB gp2 SSD disk for running local geth    
Adjust availability zone to match instance az   
```
aws --profile notation ec2 create-volume --size 500 --region us-east-1 --availability-zone us-east-1a --volume-type gp2  
aws --profile notation ec2 attach-volume --device /dev/sdh --instance-id <instance-id> --volume-id <volume-id>  
# run locally on the box:  
sudo mkfs.ext4 /dev/xvdh  
sudo mkdir /d2     
echo "UUID=<volume UUID> /d2 ext4 defaults 0 2" | sudo tee -a /etc/fstab  
sudo mount /dev/xvdh /d2   
```  

**Set the hostname**  
```
hostname tc001.mydomain.com
# add to /etc/hosts
# replace contents of /etc/hostname (with only hostname, not FQDN)
```


**Filesystem operations**   
For this setup, All LivePeer-specific files (binaries, logs, ethereum accounts, keys, etc) live on a dedicated EBS volume under /d1. The EBS volume can be backed-up via EBS snapshots and easily attached to a new instance if necessary.  
```
sudo mkdir -p /d1/livepeer/logs  
sudo mv -i livepeer_linux /d1/livepeer/bin  
sudo chown -R ubuntu:ubuntu /d1/livepeer  
```  

**Raise open filehandle limits**   
As noted in this (LivePeer FAQ](https://livepeer.readthedocs.io/en/latest/transcoding.html#faq), you can encounter the "too many open files" error when running a transcoder. As Eric notes in [this forum post](https://forum.livepeer.org/t/increase-file-limit-as-a-transcoder/170), raising the open file handle limit via pam will address this, but only for cases where you are running the livepeer node manually from an interactive session (e.g., you logged in via ssh):
from https://bugs.launchpad.net/ubuntu/+source/upstart/+bug/938669  
> PAM is intended as a user oriented library, and daemons are by definition
not users. In man limits.conf, it is clearly stated:
> 
>      Also, please note that all limit settings are set per login. They
>      are not global, nor are they permanent; existing only for the
>      duration of the session.  
See also the responses to this question about the same https://askubuntu.com/a/288534  
If you're running the LivePeer binary through non-interactive processes (upstart, systemd, etc), you need to raise the limit via a different approach (see our systemd config below). We'll go ahead and raise the limits for interactive sessions in case you want to run manually to debug, etc.  
```  
echo "ubuntu soft nofile 50000" | sudo tee -a /etc/security/limits.conf
echo "ubuntu hard nofile 50000" | sudo tee -a /etc/security/limits.conf
```  
And edit `/etc/pam.d/login` and add or uncomment the line:
```
session required /lib/security/pam_limits.so
```
You don't have to restart the system, just log out and log back in, start some long-running or background process, note its PID and then look at:
```
cat /proc/<PID>/limits 
```
to confirm the limit has been raised.  

**Install geth and run in light mode**  
geth systemd stuff  
Is a geth config file possible?  
Still need to periodically kill it ...  
```
sudo apt-get install -y software-properties-common
sudo add-apt-repository -y ppa:ethereum/ethereum
sudo apt-get update
sudo apt-get install -y ethereum  
```
geth data and logs will all live on a dedicated EBS volume (but not binaries, those get installed in default locations via apt-get install) under /d2/ for easy backups via snapshots and to easily attach to a new instance.  Run geth via systemd.  
```
sudo mkdir /d2/geth-data
sudo chown -R ubuntu:ubuntu /d2/geth-data
cd /d2/geth-data

# make sure existing .ethereum files are in place now, depends on /d2/geth-data/.ethereum  
# via systemd
sudo cp /d1/livepeer-transcoder-ops/geth/private/config/systemd ??? maybe /geth.service /etc/systemd/system/
sudo systemctl enable geth    [or reenable]
sudo systemctl start|stop|restart geth

# check the status and logs
sudo systemctl status geth ??? correct?
sudo journalctl -u geth.service -f
```  

Wait a few minutes and make sure geth is grabbing latest blocks. Sometimes you have to wait 15 minutes, kill it, and restart it before it begins syncing them.  


**Install systemd config for LivePeer**  
```  
??? If you have existing data files / keys, copy them into place now (.lpData, etc)   ??? 
??? sudo cp <path to>/livepeer-transcoder.service /etc/systemd/system/
sudo systemctl enable livepeer-transcoder    [or reenable if copying updated config]
sudo systemctl start|stop|restart livepeer-transcoder
# check status and watch the logs
systemctl status livepeer-transcoder.service
sudo journalctl -u livepeer-transcoder.service -f
```  
??? link: If you are going to use existing Ethereum accounts, go ahead and copy those into place in /d1/livepeer/.lpData/.  
Run LivePeer manually for the initial run to make sure it can:  
  * Connect to the local geth instance  
  * Detect your existing Ethereum account / keys if they are in place OR  
  * ??? Below needs link to security discussion  
  * Create a new Ethereum account if necessary. For my installation, I created the initial account *without* a passphrase for operational reasons, taking into account all the security considerations discussed here. Your mileage may vary and my recommendation is to keep security as the top priority while adjusting for your own operational environment.  
I am running LivePeer with the following params (as seen in the systemd unit config):  
**Note:** This will run as a transcoder on **mainnet**, this is basically running live in production.  
```
/d1/livepeer/bin/livepeer -datadir /d1/livepeer/.lpData -ipfsPath /d1/livepeer -log_dir /d1/livepeer/logs -ethUrl ws://127.0.0.1:8546 -v 6 -initializeRound -transcoder -publicIP 35.173.105.24 -gasPrice 0
```

??? Now run from systemd  
transfer ETH and LPT to node (small amount at first)  
Now enroll as a transcoder  


**Operational Notes**  
* Running with `initiliazeRound` is a nice thing to do - the round can't start until somebody calls it and `reward()` cannot be called until the round has been started. Running with `initializeRound` can get expensive when gas is high (I've seen ~$40)  
* Making sure `reward()` gets called everyday is the most important thing right now, after making sure everything is up and running. This generally succeeds, but, in the absence of rock-solid monitoring and alerting on this event, you should manually check it every day. Go set a reminder in your calendar to check it every day at 4pm. While you're there, set another reminder at 9pm. If the call hasn't succeeded for the day, use the command line interface to call `reward()` manually. Some reasons I've seen that can cause it to fail:  
  * You don't have enough ETH in your transcoder's account. You should monitor this and replenish as necessary.  
  * If gas prices spike, this can cause slowness and for transactions to fail, especially if you don't have enough funds (see above).  
  * Unable to communicate with the geth node - I've seen the local geth node appear to run fine and continue to stay sync'd to latest blocks and log that it's submitting transactions (such as calls to reward), but they fail silently and no errors or warnings are produced. LivePeer [issue #455](https://github.com/livepeer/go-livepeer/issues/455) documents a problem similar to this. In such cases, I've restarted first the geth node, waited for it to sync (a couple minutes at most), and then restarted the livepeer node. This is annoying enough to consider restarting geth automatically on a nightly (!) basis.  
* What is the best way to backup the account / credentials tied to the node?  
* How can you migrate your transcoder to different hardware but maintain the same transcoder account and ID?  
* What livepeer / ipfs / etc logs needs to be rotated?  
* Keep an eye on the LivePeer [Releases](https://github.com/livepeer/go-livepeer/releases) page for updates to the software, as well as the [discord](https://discord.gg/cBfD23u) and [forum](https://forum.livepeer.org/) discussions.  


**LivePeer questions I had but was able to answer**  
* **Note** Don't forget the [upcoming networking updates!](https://forum.livepeer.org/t/upcoming-networking-upgrades/298)  
* The most complicated part is knowing the correct steps and order to take in the CLI to make sure the transcoder is active and how to debug if it isn't, also what options to start it with. There isn't an official walkthrough of recommended arguments to start LP with and then register on mainnet as a transcoder.  
* Is it ok to call `reward()` more than once per round? Yes, it will just say "reward already called for this round."  
* Is it worth setting up a dedicated ipfs node in local network? Doesn't look like it's necessary at this time.  
* GPU - Is it worth it to run with GPU? How much does it help? What specifically leverages the GPU - ffmpeg? short answer: Not yet  
  * Adding GPU Acceleration to transcoding is still an [open issue](https://github.com/livepeer/lpms/issues/33). 
  * GPU transcoding is not currently supported, according to Doug, "Currently we support deterministic CPU transcoding, but we're working on what you read in the above proposal to enable GPU transcoding in a way that will not disrupt GPU mining operations"  
  * In [issue #51 Transcoder Design](https://github.com/livepeer/lpms/issues/51#issuecomment-362502511), j0sh goes into a bit more depth on which areas may benefit from GPU    
  > There are some workloads in the transcoding pipeline that might benefit from GPU (such as colorspace conversion), but encoding generally benefits more from SIMD (AVX) or fixed function hardware (QuickSync). That being said, FFMpeg already supports the Intel MediaSync SDK which I believe is able to run certain operations on the (Intel?) GPU natively. I'm hoping that enabling MediaSync support is as simple as installing the library and setting the ffmpeg configure flag. We'd likely need run-time hardware detection as well.
  > GPU's might help more with verification, but it'd depend on the method we choose.   
  * See also the [Transcoder Design doc](https://github.com/livepeer/lpms/wiki/Transcoder-Design)  
  * There is a [GPU transcoding verficiation proposal](https://github.com/livepeer/research/issues/12) in [research projects](https://github.com/livepeer/research/projects/1#card-9975184)    
  * When GPU's can be meaningfully helpful, [P2 GPU instances](https://aws.amazon.com/ec2/instance-types/p2/) are one (very expensive) option, or [Elastic GPUs](https://aws.amazon.com/ec2/elastic-gpus/details/), which can be attached to certain instance types.  
* What do you need to do to transfer your transcoder identity to a new box? eg if you need to migrate hardware for some reason?  
  * I guess the identity is just the eth address of the account, so as long as you migrate that to a new machine it should be fine    
* Difference between reward, stake, pending stake, etc. 


**LivePeer open questions**  
* How to know if you've been slashed?  
* Specifying `-log_dir` on the command line only moved where the ipfs log file got written, `livepeer` still wrote its log to stderr.  
* Is it possible to transfer LPT from a transcoder to another account without `unbonding()` the entire stake? Is this done via CLI option #11 "transfer" or can you unbond (a certain number) and then call "withdraw stake" on just that portion?  
* What ports should be open to internal network? Open to the world?  
* Gas: Doug says 10Gwei is a safe price - does that mean you’ll pay 10Gwei every time?? or that’s just max price  
* How much ETH should you keep in your account?  
* Capacity planning - how to estimate transcoding rate (how long to transcode each second of output video) based on machine resources?  
* Unclear from docs: needs ffmpeg? the specially built static version? https://github.com/livepeer/ffmpeg-static  
* How can you run multiple transcoder instances, behind a load balancer, for example, but have them all use the same identity? Because you just register as a single transcoder id, right?   



**Becoming an active transcoder on mainnet**  
* **spin up a fresh node but try to put an old account in place before starting the livepeer binary**  
* Fund your node with ETH and LPT and bond to yourself  
* Specifying the Ethereum account - Eth Account Each Livepeer node should have an Ethereum account. Use -ethAccountAddr to specify the account address. You should make sure the keys to the account is in the keystore directory of ethDatadir you passed in.  



**HTTP Query interface**  
* `curl http://localhost:8935/getAvailableTranscodingOptions`  
* Can also set broadcast config by POST'ing params to http://localhost:8935/setBroadcastConfig   


**Future Architecture Directions**    
  


**Reference**   
  * Master reference docs and info is aggregated in this thread - [Transcoder Megathread - Start here to learn about playing the role of transcoder on Livepeer](https://forum.livepeer.org/t/transcoder-megathread-start-here-to-learn-about-playing-the-role-of-transcoder-on-livepeer/190)   

**OPs TODO**  
- Configuration  
  - Use actual config management  
  - Testing  
  - Automated deployment  
  - Docker?  
- Traffic management  
  - Load balancing  
  - Automatic failover  
  - Regional routing  
  - Auto-scaling  
- Monitoring, Alerting, Metrics Collection  
  - Health checks of LivePeer instance  
  - Monitor and alert if reward() doesn't get called  
  - Monitor amount of ETH in transcoder's account and alert if below certain threshold.  
- Security  
  - Better management of Ethereum private keys  
  - Possibly using Hashicorp's Vault for private keys or AWS KMS   
- EBS Volumes
  - Automate EBS snapshots  
  - Encrypt EBS Volumes by default?  
- Add GPU's, optionally  
- Local geth node  
  - Would it benefit from being a fast-sync node or a full node?  
  - Should probably move geth to a dedicated instance that multiple local transcoder nodes can connect to  
  - Should probably run a local geth cluster in each region you plan to run transcoders  
- Log rotation for LivePeer and geth logs    
- Helpful to give the instance an DNS and/or ENS name?  
- Better documentation of AWS Security groups, IAM users and permissions, ssh gateway host, etc  


**Notes to move elsewhere**  
* yep, this is the port to poll to track status and info. Available commands are documented in [webserver.go](https://github.com/livepeer/go-livepeer/blob/ec288f43b60fbf3bd61f81b636538b5b004aaa86/server/webserver.go)  
* Seems like it can send server metrics to http://viz.livepeer.org:8081/metrics ? see [livepeer.go](https://github.com/livepeer/go-livepeer/blob/master/cmd/livepeer/livepeer.go) interesting that it can end metrics by default, wonder if that can be redirected and to what kind of server. you can also specify the monitor host to send to  
  * Maybe to the monitor server here? https://github.com/livepeer/go-livepeer/blob/master/monitor/monitor.go  
  * Looks like there's a separate monitor server project https://github.com/livepeer/streamingviz  ... although it hasn't been touched in a year  