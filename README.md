## Running a LivePeer transcoder   
The purpose of this project is to document my own approach to running a LivePeer transcoder in production. The goal is to run robust infrastructure for the LivePeer transcoding network and to share any supporting code or processes to help the community do the same. These are the early steps in building a robust operating framework in which to run a transcoder network. Some of the operational characteristics I'm working toward include:  
  * Availability (including fast recovery)  
  * Security  
  * Flexibility  / composability  
  * Repeatability  
  * Capacity understanding not the same as performance  
  * Configuration  / Config is code   

## Key Decisions  
Some key decisions I made and why.  
- **Platform** - AWS, Linux, Ubuntu - addressable via API, flexibility, elastic capacity.  
- **Hardware Resources** - initially we're over-provisioned to make sure we have plenty of capacity.  
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
- **Addressability** - fixed ip address (AWS Elastic IP Address) for stable addressability, ability to moved ip between instances for better flexibility.  
- **Storage performance** - gp2 SSD for decent, consistent performance.  
- **Storage flexibility** - EBS Volumes - concentrate data and config with deliberate filesystem location choices, not in default home directories, on dedicated volumes separate from root disk, flexilibity - easily expandable, easily transferred to new instance (speed of recovery), easy to backup (EBS snapshots, easily automated).  
- **Process supervision** - systemd (ugh). Not a huge fan of systemd, but given that it's default now, there's a lot of value in not fighting the native system.  
- **Timekeeping** is always crucial. In Ubuntu 18.04, the base system uses [systemd-timesyncd](https://www.freedesktop.org/software/systemd/man/timedatectl.html) which looks ok, but may want to consider using [Chrony](https://chrony.tuxfamily.org/) for more fine-grained control of accuracy and syncing. See the [FAQ](https://chrony.tuxfamily.org/faq.html), this [Ubuntu help article on time sync](https://help.ubuntu.com/lts/serverguide/NTP.html), and a [basic overview of configuring chrony](https://blog.ubuntu.com/2018/04/09/ubuntu-bionic-using-chrony-to-configure-ntp).  
- **Ethereum node access** - we run a local (light) geth node  
  * The official docs [recommend running geth](https://livepeer.readthedocs.io/en/latest/node.html) but other info, such as [this forum post](https://forum.livepeer.org/t/how-to-run-livepeer-with-geth/143), say it's not necessary. That's correct, it's not strictly required but I think it's clearly beneficial - as mentioned in this [FAQ](https://livepeer.readthedocs.io/en/latest/transcoding.html#faq), a flaky connection to the Ethereum network can lead to errors, calls to `reward()` failing, failure to transcode incoming jobs, etc. The best way to ensure a solid, fast connection to the Ethereum network is to run a local geth / parity node.   
  * See this post for args to run a local geth instance https://forum.livepeer.org/t/transcoder-tip-geth-light-client/247/7  
  * Need a full copy of ETH blockchain? It seems a fast sync is sufficient  
  * My preference is to run it on a dedicated local node (not the transcoder)  
- **Security**  
This is not meant to be an exhaustive review of security practices, just a quick overview of some considerations that are top of mind.  
* Securing Ethereum keys is a key consideration. Automating startup of `livepeer` by automatically supplying a password for the ethereum?  
  * I just supplied a blank password the first time and then it doesn't ask for password on startup in the future  
  * What are implications of no password? For backing up files, for security of account in general, etc?  
  * Could just do it via command-line, but don't really want it to be visible to 'ps'  
  * Would prefer at least a config file if nothing else ... 
* Ports - all locked down, closed to the world and the local network except 4433, as required by [upcoming network updates](https://forum.livepeer.org/t/upcoming-networking-upgrades/298) which I think has to be open to the world.
  * Note that ssh is also closed to the world and, in our setup, only accessible through an ssh bastion host which runs ssh on a non-standard port and is locked down except to specific, known IPs.    
* No root logins, auth via ssh keys only, keep security patches up-to-date, review all running procs and open ports, shut down (permanently) all unnecessary ones, make sure you have 2FA enable for your AWS account, backup regularly and automatically, monitor your boxes, regularly audit and rotate authorized ssh keys, AWS IAM permissions, sudo access, don't allow access via root AWS ssh keys, encourage audit trails via access via username-accounts (vs system accounts such as "ubuntu"), etc. Be aware of security implications of backups, esp snapshots of volumes  
* Securing your node and access to private ETH key  
  * Tradeoffs, hacks, etc.  
  * Storing private key in [AWS Parameter Store](https://aws.amazon.com/systems-manager/features/#Parameter_Store) in AWS [Key Management Service](https://aws.amazon.com/kms/)?  
  * [Getting started with AWS Parameter store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html) see also How AWS [Systems Manager Parameter Store Uses AWS KMS](https://docs.aws.amazon.com/kms/latest/developerguide/services-parameter-store.html?shortFooter=true)  
  * goog: [golang example aws kms](https://www.google.com/search?biw=1295&bih=1103&ei=N30yW4aHJ8vxzgLp1J-gBw&q=golang+example+aws+kms&oq=golang+example+aws+kms&gs_l=psy-ab.3..33i22i29i30k1.231231.237437.0.238301.22.17.0.4.4.0.305.1360.3j7j0j1.11.0....0...1c.1.64.psy-ab..8.14.1249...0j0i67k1j0i131i67k1j0i131k1j0i22i30k1.0.7Ap2nvkZiVw)  
  * It's just that the ethereum client doesn't seem to have the capability to get the account key over the network or from anything other than a file https://github.com/livepeer/go-livepeer/blob/master/eth/accountmanager.go  


  
**Note:** This is all very specific to AWS and Ubuntu. I haven't done the work to generalize for Amazon Linux, RHEL, CentOS, whatever.   

* Future Architecture Directions (see OPs TODO)  
There is much room for improvement. See below for some specific areas of known technical debt and future work.  


**Instance type and resources**  
I want to be sure this transcoder can perform, so for the initial phase I've overprovisioned the resources of CPU, RAM, disk performance, and bandwidth (details below). This means this specific configuration is expensive so feel free to choose lower-resource instance types.  


You can use the [AWS Command Line Interface](https://docs.aws.amazon.com/cli/latest/userguide/installing.html) to launch instances with these characteristics using [this configuration file](https://github.com/alexlines/livepeer-transcoder-ops/blob/master/private/config/aws-instances/livepeer-transcoder-ec2-config.json) as follows:  
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


Grab latest LivePeer binaries  
You can build from scratch if you want but why ... I won't go into that, read more about it in the [official README](https://github.com/livepeer/go-livepeer/blob/master/README.md)    
Download the latest mainnet-targeted livepeer and livepeer_cli from https://github.com/livepeer/go-livepeer/releases.  
```
cd
curl -s -L https://github.com/livepeer/go-livepeer/releases/download/0.2.4/livepeer_linux.tar.gz > livepeer_linux.tgz
gzip -d -c livepeer_linux.tar.gz | tar xvf -
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
sudo mv -i ~/livepeer_linux /d1/livepeer/bin  
sudo chown -R ubuntu:ubuntu /d1/livepeer  
cd /d1
git clone git@github.com:alexlines/livepeer-transcoder-ops.git
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
??? Still need to periodically kill it ...  
```
sudo apt-get install -y software-properties-common
sudo add-apt-repository -y ppa:ethereum/ethereum
sudo apt-get update
sudo apt-get install -y ethereum  
```
geth data, logs, and any keys (but not binaries, those get installed in default locations via apt-get install) will all live on a dedicated EBS volume under /d2/ for easy backups via snapshots and to easily attach to a new instance.  
Setup geth data directories on the attached EBS volume and copy systemd unit file into place:  
```
sudo mkdir /d2/geth-data
sudo chown -R ubuntu:ubuntu /d2/geth-data
sudo cp /d1/livepeer-transcoder-ops/private/config/geth/systemd/geth.service /etc/systemd/system/
sudo cp /d1/livepeer-transcoder-ops/private/config/geth/geth-config.toml /d2/geth-data/
```
If you plan to use existing .ethereum files or keys, copy them into place now in `/d2/geth-data/.ethereum`  
start geth via systemd and watch the logs:
```
sudo systemctl enable geth    [or 'reenable' if you're overwriting existing config file]
sudo systemctl start|stop|restart geth

# check the status and logs
sudo systemctl status geth
sudo journalctl -u geth.service -f
```  

Wait a few minutes and make sure geth is grabbing latest blocks. Sometimes you have to wait 15 minutes, kill it, and restart it before it begins syncing them.  


**Install systemd config for LivePeer**  
If you are going to use existing LivePeer account data, go ahead and copy it into place now in `/d1/livepeer/.lpData/`  
```  
sudo cp /d1/livepeer-transcoder-ops/private/config/livepeer/systemd/livepeer-transcoder.service /etc/systemd/system/
sudo systemctl enable livepeer-transcoder    [or reenable if copying updated config]
```  
  
Run LivePeer manually for the initial run to make sure it can:  
  * Connect to the local geth instance  
  * Detect your existing Ethereum account / keys if they are in place **OR**  
  * Create a new Ethereum account if necessary. For my installation, I created this initial account *without* a passphrase for operational reasons, taking into account all the security considerations discussed elsewhere in this document. Your mileage may vary and my recommendation is to keep security as the top priority while adjusting for your own operational environment.  

**Running live on the Ethereum mainnet**  
* Transfer some ETH and LPT to your node - a small amount at first to confirm addresses and process.  
* Run the livepeer binary by hand as an initial test. I am running LivePeer with the following params (as seen in the [systemd unit config](https://github.com/alexlines/livepeer-transcoder-ops/blob/master/private/config/livepeer/systemd/livepeer-transcoder.service)):  
**Note:** This will run as a transcoder on **mainnet**, this is basically running live in production.  
```
/d1/livepeer/bin/livepeer -datadir /d1/livepeer/.lpData -ipfsPath /d1/livepeer -log_dir /d1/livepeer/logs -ethUrl ws://127.0.0.1:8546 -v 6 -initializeRound -transcoder -publicIP <public ip> -gasPrice 0
```

Now kill that process and start livepeer using systemd and watch the logs:
```
kill $(pgrep livepeer)
sudo systemctl start livepeer-transcoder
# check status and watch the logs
sudo systemctl status livepeer-transcoder
sudo journalctl -u livepeer-transcoder.service -f
```

Now use the livepeer command line utility to enroll as a transcoder and set transcoder config:  
```
/d1/livepeer/bin/livepeer_cli
```
Choose `13. Invoke multi-step "become a transcoder"`  



**Operational Notes**  
* Running with `initiliazeRound` is a nice thing to do - the round can't start until somebody calls it and `reward()` cannot be called until the round has been started. Running with `initializeRound` can get expensive when gas is high (I've seen ~$40)  
* Making sure `reward()` gets called every day is the most important thing right now, after making sure everything is up and running. This generally succeeds, but, in the absence of rock-solid monitoring and alerting on this event, you should manually check it every day. Go set a reminder in your calendar to check it every day at 4pm. While you're there, set another reminder at 9pm. If the call hasn't succeeded for the day, use the command line interface to call `reward()` manually. Some reasons I've seen that can cause it to fail:  
  * You don't have enough ETH in your transcoder's account. You should monitor this and replenish as necessary.  
  * If gas prices spike, this can cause slowness and for transactions to fail, especially if you don't have enough funds (see above).  
  * Unable to communicate with the geth node - I've seen the local geth node appear to run fine and continue to stay sync'd to latest blocks and log that it's submitting transactions (such as calls to reward), but they fail silently and no errors or warnings are produced. LivePeer [issue #455](https://github.com/livepeer/go-livepeer/issues/455) documents a problem similar to this. In such cases, I've restarted first the geth node, waited for it to sync (a couple minutes at most), and then restarted the livepeer node. This is annoying enough to consider restarting geth automatically on a nightly (!) basis.    
* What is the best way to backup the account / credentials tied to the node? For this setup, just create a snapshot of the EBS volume.  
* What livepeer / ipfs / etc logs needs to be rotated? It looks like the only logs that livepeer currently writes are ipfs logs which, in this configuration, are written to /d1/livepeer/logs/ and are automatically rotated every 10MB, though not automatically compressed.  
* Keep an eye on the LivePeer [Releases](https://github.com/livepeer/go-livepeer/releases) page for updates to the software, as well as the [discord](https://discord.gg/cBfD23u) and [forum](https://forum.livepeer.org/) discussions.  


**LivePeer questions I had but was able to answer**  
* **Note** Don't forget the [upcoming networking updates!](https://forum.livepeer.org/t/upcoming-networking-upgrades/298)  
* Is it ok to call `reward()` more than once per round? Yes, I think it will just say "reward already called for this round."  
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
* What do you need to do to transfer your transcoder identity to a new box? eg if you need to migrate hardware for some reason? The identity of the transcoder node is just the eth address of the account, so as long as you migrate that to a new machine it should be fine - backup and restore your livepeer .lpData directory (but it's very sensitive and contains your private key, so be very careful how and where you back it up, encrypt it, limit access, make sure it's not a publicly accessible machine, storage account, etc.  
* Explain difference between reward, stake, pending stake, etc. 


**LivePeer open questions**  
* How to know if you've been slashed?  
* Specifying `-log_dir` on the command line only moved where the ipfs log file got written, `livepeer` still wrote its log to stderr.  
* Is it possible to transfer LPT from a transcoder to another account without `unbonding()` the entire stake? Is this done via CLI option #11 "transfer" or can you unbond (a certain number) and then call "withdraw stake" on just that portion?  
* What ports should be open to internal network? Open to the world?  
* How much ETH should you keep in your transcoder account?  
* Capacity planning - how to estimate transcoding rate (how long to transcode each second of output video) based on machine resources?  
* How can you run multiple transcoder instances, behind a load balancer, for example, but have them all use the same identity? Because you just register as a single transcoder id, right? Pretty sure it's not yet possible.  
* How to monitor demand in different regions? Would be great to know that there is more demand than capacity in Asia/Pacific, for example, and to spin up capacity there.  



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
  - Go and systemd both support watchdog for process health monitoring http://0pointer.de/blog/projects/watchdog.html  
  - It would be nice if the livepeer internal webserver supported a call to `/health` for example, which could be checked by a nagios/etc plugin, as well as by a load balancer which was monitoring the health of a pool of transcoders.  You could sortof fake this today by using a call to `http://<transcoder ip>:8935/nodeID` and make sure it returns the expected nodeID, but a proper `/health` function could better check a few vital signs.  
  - Monitor and alert if reward() doesn't get called. A few ways to monitor this:  
    - Query the local livepeer node via http:  
    ```
      $ curl http://127.0.0.1:8935/transcoderInfo
      {"Address":"0x50d69f8253685999b4c74a67ccb3d240e2a56ed6","LastRewardRound":1018,"RewardCut":30000,"FeeShare":300000,"PricePerSegment":150000000000,"PendingRewardCut":30000,"PendingFeeShare":300000,"PendingPricePerSegment":150000000000,"DelegatedStake":6454553077282307328907,"Active":true,"Status":"Registered"}
      ```  
     and if `LastRewardRound` doesn't match the current round (which you have to get via another call), then `reward()` has not yet been called. This would be straightfoward to monitor automatically and you could alert on this after a certain time of day.     
    - The highest certainty check would be to query the Ethereum blockchain for the tx where your node called `reward()` 
      - Could do this by querying the local geth node via [JSON-RPC api](https://github.com/ethereum/wiki/wiki/JSON-RPC), e.g. using [eth_getTransactionbyHash](https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_gettransactionbyhash):  
      ```
      curl -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionByHash","params":["0xcde8ec889fa7ed433d2a55c5f34f1be98f4dad97791a27c258d18eb1bad17d0f"],"id":1}' http://localhost:8545  
      ```
      but there's not an easy way to list recent transactions for an account or contract ... looks like filters/logs are the way to do this? https://github.com/ethereum/go-ethereum/issues/1897  or here https://github.com/ethereum/go-ethereum/issues/2104     
      - Use the [etherscan API](https://etherscan.io/apis), the `input` value below indicates a call to `reward()`:    
      ```
      $ curl 'http://api.etherscan.io/api?module=account&action=txlist&address=<your node address>&startblock=5858336&endblock=5858337&sort=desc&apikey=<api key>'
      {
      "status":"1",
      "message":"OK",
      "result":[
        {
        "blockNumber":"5944652",
        "blockHash":"0x58d1fcc7ec68dd001a3c166572b3a2d308f4b0598a92faf46e509abb70118de3",
        "timeStamp":"1531311107",
        "hash":"0x5960471df2da7cc405a4bcb5a195c73a711b37ada828a56672b4ef40c891b918",
        "nonce":"203",
        "transactionIndex":"136",
        "from":"0x345551571c5ef20111c6168b9a498dfb836e7c09",
        "to":"0x511bc4556d823ae99630ae8de28b9b80df90ea2e",
        "value":"0",
        "gas":"262287",
        "gasPrice":"8000000000",
        "input":"0x228cb733",
        "contractAddress":"",
        "cumulativeGasUsed":"3474053",
        "txreceipt_status":"1",
        "gasUsed":"213639",
        "confirmations":"1143",
        "isError":"0"
        }]
      }
      ```  
  - Monitor amount of ETH in transcoder's account and alert if below certain threshold.  
  - Metrics collection - one idea is [publish metrics to AWS CloudWatch](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/publishingMetrics.html), possibly with [custom events](https://aws.amazon.com/blogs/security/how-to-use-amazon-cloudwatch-events-to-monitor-application-health/)? Though I haven't researched if feasible in this case.  
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


