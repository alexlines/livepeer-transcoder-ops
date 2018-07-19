package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"

	"github.com/golang/glog"
	"github.com/ethereum/go-ethereum/common"

)

// these types and funcs are copied or based on files from go-livepeer, mostly
// from files in github.com/livepeer/go-livepeer/cmd/livepeer_cli
// This is a quick proof of concept, they should more properly be imported 
// and used as a base to build on.

// The initial logic is that if the currentRound and LastRewardRound are not 
// the same, then reward() has not yet been called for this round. 
// A better way to check this:
// - has current round been initialized
// - what is current round length
// - what is start block for current round
// - how many blocks to wait until next round
// - alert based on configurable threshold if reward() has not been 
//   called and there are less than n blocks until next round

// Think also about what action a responding operator can take in this case,
// is it worth paging someone at 3am? What can the script attempt to 
// remedy the situation?

// TODO's also include:
// - accept command lines args (e.g. where does livepeer server live?)
// - should alert if reward() has not been called, subject to thresholds, etc


type Transcoder struct {
	Address                common.Address
	// tech debt:
	//  changed type of LastRewardRound
	//  from big.Int to int to simplify
	//  comparison to currentRound
	//  but should be changed back to 
	//  big.Int at some point
	//LastRewardRound        *big.Int
	LastRewardRound        *int
	RewardCut              *big.Int
	FeeShare               *big.Int
	PricePerSegment        *big.Int
	PendingRewardCut       *big.Int
	PendingFeeShare        *big.Int
	PendingPricePerSegment *big.Int
	DelegatedStake         *big.Int
	Active                 bool
	Status                 string
}

type wizard struct {
	endpoint   string // Local livepeer node
	httpPort   string
	host       string
}


func main() {
	// tech debt: hard-coding these values:
	lp_host := "localhost"
	lp_port := "8935"

	w := &wizard{
		endpoint: fmt.Sprintf("http://%v:%v/status", lp_host, lp_port),
		httpPort: lp_port,
		host:     lp_host,
	}
	w.run()
}

func (w *wizard) run() {
	// Make sure there is a local node running
	_, err := http.Get(w.endpoint)
	if err != nil {
		glog.Errorf("Cannot find local node. Is your node running on http:%v?", w.httpPort)
		return
	}

	nodeid := w.getNodeID()
	currentRound := w.currentRound()
	t, err := w.getTranscoderInfo()
	if err != nil {
		glog.Errorf("Error getting transcoder info: %v", err)
		return
	}

	// if the currentRound and LastRewardRound are not the same, then
	// reward() has not yet been called for this round. 
	if strconv.Atoi(currentRound) != strconv.Atoi(t.LastRewardRound) {
	    fmt.Printf("reward has not been called for current round %v\n", currentRound)
	    // possibly alert here, based on configurable thresholds
	}
	// another option is to return true or false
	// fmt.Printf("%v\n", (a != b))

	// don't be so chatty in the future, but debugging for now:
	fmt.Printf("current round    : %v\n", currentRound)
	fmt.Printf("nodeid           : %v\n", nodeid)
	fmt.Printf("Status           : %v\n", t.Status)
	fmt.Printf("Active           : %v\n", t.Active)
	fmt.Printf("Last Reward Round: %v\n", t.LastRewardRound.String())

}

func (w *wizard) getNodeID() string {
	return httpGet(fmt.Sprintf("http://%v:%v/nodeID", w.host, w.httpPort))
}

func httpGet(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		glog.Errorf("Error sending HTTP GET: %v")
		return ""
	}

	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil || string(result) == "" {
		return ""
	}
	return string(result)

}

func (w *wizard) currentRound() string {
	return httpGet(fmt.Sprintf("http://%v:%v/currentRound", w.host, w.httpPort))
}

func (w *wizard) getTranscoderInfo() (Transcoder, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v:%v/transcoderInfo", w.host, w.httpPort))
	if err != nil {
		return Transcoder{}, err
	}

	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Transcoder{}, err
	}

	var tInfo Transcoder
	err = json.Unmarshal(result, &tInfo)
	if err != nil {
		return Transcoder{}, err
	}

	return tInfo, nil
}
