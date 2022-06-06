package extract

import (
	//"fmt"
	"github.com/piux2/gnobounty7/sink"
)

// Extract votes from a proposal, with a given height range
// If either start or end out of the range of the heights between
// proposal submitted and tallied or dropped, it returns results within the boundry

type VoteExtract struct {
	Height   int64  `json:"height"`
	TxHash   string `json:"tx_hash"`
	Sender   string `json:"voter_address"`
	Proposal uint64 `json:"proposal_id"`
}

//if sink is true, store the result to the sink others print it as json line.
func ExtractVotes(psink *sink.PsqlSink, sink bool) error {

	if sink == false {

		return nil
	}
	psink.Config.SinkOn = sink

	keywords := []string{"MsgVote", "vote", "submit_proposal", "active_proposal", "inactive_proposal", "signal_proposal", "proposa_result", "proposal_dropped", "proposal_passed", "proposal_rejected", "proposal_failed"}

	psink.Batch.Statement = "INSERT INTO keyword_event(chain_id,height,tx_id, event_id, attribute_id,composite_key,type,attribute, value, tx_hash, keyword) VALUES "
	err := psink.SinkAbciResponses(keywords)

	return err

}
