package sink

import (
	"encoding/json"
	"fmt"

	"strconv"
	"strings"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmstate "github.com/tendermint/tendermint/proto/tendermint/state"
	tmtypes "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	"github.com/piux2/gnobounty7/config"

	"github.com/piux2/gnobounty7/psql"
	"github.com/piux2/gnobounty7/state"
	"github.com/piux2/gnobounty7/store"
	putil "github.com/piux2/gnobounty7/util"

	"github.com/golang/protobuf/proto" // used by tendermint
)

// Find Keyword and Sink it
type PsqlSink struct {
	BlockStore *store.BlockStore
	StateStore *state.Store

	AppStore         *storetypes.CommitMultiStore
	AppDB            tmdb.DB
	EventSink        *psql.EventSink // postgres sql database sink
	Top              int64           //top height
	Base             int64           // bottom height
	Batch            InsertBatch     // number of events inserted in psql in a batch
	Config           *config.Config
	isProposalReady  bool
	isValidatorReady bool
}

func NewPsqlSink(c *config.Config) *PsqlSink {

	bs, ss, err := LoadStateAndBlockStore(c)

	if err != nil {
		fmt.Println("load state.db and blockstore.db failed", err)
		panic(err)
	}

	appdb := LoadAppDB(c)

	as := LoadAppStore(appdb)

	es, err := loadPsqlSink(c)
	if err != nil {

		panic(err)

	}

	batch := InsertBatch{batchSize, "", []string{}, []interface{}{}, ""}

	return &PsqlSink{BlockStore: bs, StateStore: &ss, AppStore: as, AppDB: appdb, EventSink: es, Batch: batch, Config: c}
}

func loadPsqlSink(cfg *config.Config) (*psql.EventSink, error) {
	fmt.Printf("Load postgres... ")

	conn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.PsqlHost, cfg.PsqlPort, cfg.PsqlUser, cfg.PsqlPassword, cfg.PsqlDBName,
	)
	if conn == "" {
		return nil, fmt.Errorf("the psql connection settings cannot be empty")
	}

	es, err := psql.NewEventSink(conn, cfg.ChainID)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Done \n")
	return es, nil
}

func (psink *PsqlSink) Close() {

	psink.BlockStore.Close()
	(*psink.StateStore).Close()
	psink.EventSink.Stop()

}

// Assumption: Finding individual events first before insert them to the databse is a lot faster
// than dump everyting to psql, since the number of events are lot smaller comparing to
// entire KV store. and write to psql is much more time consuming
// Also since certain information  are inside value. for example MsgVote options.

// find the kv store and insert it to the datbase for certain message type

func (psink *PsqlSink) SinkAbciResponses(keyword []string) error {

	keyPrefix := "abciResponsesKey:"

	db := (*psink.StateStore).DB()
	//tdb := db.(*tmdb.GoLevelDB)
	//ldb := tdb.DB()
	var counter int64 = 0

	for i := psink.Base; i <= psink.Top; i++ {
		//for i := int64(8625971); i <= int64(8625971); i++ {

		key := []byte(keyPrefix + strconv.FormatInt(i, 10))

		value, err := db.Get(key)

		if err != nil {
			fmt.Printf("At Height %d, failed to get abciResponsesKey %s \n", i, err)
			continue
		}

		r := tmstate.ABCIResponses{}

		if err := proto.Unmarshal(value, &r); err != nil {

			return fmt.Errorf("At height %d failed to parse response value %v", i, err)
		}

		if err := psink.findResponsesEvents(&r, i, keyword); err != nil {

			return fmt.Errorf(">At height %d failed to find events %v ", i, err)
		}

		counter++
		fmt.Printf("\r processing response %s, processed %d ", key, counter)
	}
	fmt.Printf("\n total processed %d blocks", counter)
	// flush the remaining unexected statement
	err := psink.ExecBatch()
	if err != nil {
		panic(err)
	}

	return nil
}

func matchKeywords(es []abcitypes.Event, keywords []string) (string, bool) {

	//func matchKeywords(tx *abcitypes.ResponseDeliverTx, keywords []string) (string, bool) {

	var keyword string = ""
	var hasKeyword bool = false
	var ok bool = false

	for _, e := range es {

		keyword, ok = putil.StringContains(e.Type, keywords)

		if ok == true {
			putil.PrettyJson(e)
			hasKeyword = true
			break
		}

		for _, a := range e.Attributes {

			keyword, ok = putil.StringContains(a.Key, keywords)

			if ok == true {
				putil.PrettyJson(e)
				hasKeyword = true
				break
			}

			keyword, ok = putil.StringContains(a.Value, keywords)

			if ok == true {
				putil.PrettyJson(e)
				hasKeyword = true
				break
			}

		}

		if hasKeyword == true {
			break
		}
	}

	return keyword, hasKeyword
}

func (psink *PsqlSink) findResponsesEvents(r *tmstate.ABCIResponses, height int64, keywords []string) error {
	// blockstore.block.data.Txs[i] <==> statestore.abciResponses.DeliverTxs[i] the sequence are ordered, mapped by index
	// in the same hight.
	// Reference: https://github.com/tendermint/tendermint/blob/8682489551b69d6b31947c9253c8c2f86fe4f2c7/cmd/tendermint/commands/reindex_event.go#L195

	k, ok := matchKeywords(r.EndBlock.Events, keywords)

	if ok == true && k != "" {
		fmt.Printf("proposal tallied matched %s in end block %d\n", k, height)

		b := psink.BlockStore.LoadBlock(height)
		blockHash := b.Hash().String()

		for i, e := range r.EndBlock.Events {
			err := psink.insertTxEvent(height, -1, i, blockHash, k, e)

			if err != nil {
				return err
			}
		}

		//		putil.PrettyJson(r.EndBlock.Events)

	}

	for txIndex, tx := range r.DeliverTxs {

		// filter out unsucessful transaction
		if tx.Code != 0 {

			continue

		}

		hasKeyword := false
		keyword := ""

		keyword, hasKeyword = matchKeywords(tx.Events, keywords)

		if hasKeyword == true && keyword != "" {
			if psink.Config.SinkOn == true {
				err := psink.insertTx(height, txIndex, tx, keyword)
				if err != nil {

					return err
				}
			} else {

				//		fmt.Println(keyword)
				fmt.Printf("proposal sumited matched %s in end block %d\n", keyword, height)
				//	putil.PrettyJson(tx)

			}

		}
	}

	return nil
}

type EventsLog struct {
	Events []abcitypes.Event `json:"events"`
}

func (psink *PsqlSink) insertTx(height int64, txIndex int, tx *abcitypes.ResponseDeliverTx, keyword string) error {

	txHash, err := psink.getTxHash(height, txIndex)
	if err != nil {

		return err
	}

	// There are fees events in the MsgVote Transactions
	// we do not need to include it in the sql database.
	// transaction log does not include fee transactions

	elogs := []EventsLog{}

	if err := json.Unmarshal([]byte(tx.Log), &elogs); err != nil {
		fmt.Printf("txLog: %v\n", tx.Log)
		return fmt.Errorf("\n> Failed to parse tx.Log  %v", err)
	}

	for eventIndex, e := range elogs[0].Events {

		err = psink.insertTxEvent(height, txIndex, eventIndex, txHash, keyword, e)
		if err != nil {
			return err
		}

	}
	return nil
}

// provide the height of block in block store and i-th tx in the block.
func (psink *PsqlSink) getTxHash(height int64, txIndex int) (string, error) {
	i := txIndex

	b := psink.BlockStore.LoadBlock(height)
	if b == nil {
		return "", fmt.Errorf("\n> not able to load block at height %d from the blockstore", height)
	}
	l := len(b.Data.Txs)
	if i > l {

		return "", fmt.Errorf("\n> out of bound for at height %d from the blockstore: txs length %d (i=%d) ", height, l, i)
	}
	tx := b.Data.Txs[i]
	txHash := fmt.Sprintf("%X", tmtypes.Tx(tx).Hash())

	return txHash, nil

}

type KeywordEvent struct {
	chainId    string
	height     int64
	txIndex    int
	eventIndex int
	attriIndex int

	compositeKey string

	eventType  string
	attriKey   string
	attriValue string

	txHash  string
	keyword string
}

// We use InsertBatch to pack the value argemnts to insert multi rows in one insert statement
// the length of valueArgs should equal to length of  number of argment per row x number of rows
type InsertBatch struct {
	Size         int           // the batch limit
	ValueStr     string        // (?,?,?..) indicates number of value per row.
	ValueStrings []string      //  store and append one '(?,?,?...)' per row
	ValueArgs    []interface{} // store and append every arguments in a row
	Statement    string
}

//
func NewInsertBatch(size int, valueStr string, valueStrings []string, args []interface{}, stmt string) InsertBatch {

	return InsertBatch{

		Size:         batchSize,
		ValueStr:     valueStr,
		ValueStrings: valueStrings,
		ValueArgs:    args,
		Statement:    stmt,
	}

}

// a good size on a regular pc laptop
const batchSize int = 200

// convert []{(?,?,?), (?,?,?), (?,?,?)}to "($1,$2,$3),($4,$5,$6),($7,$8,$9) for sql prepared statement parameters
func (psink *PsqlSink) prepareValueStrings() string {

	s := []string{}

	counter := 0

	for _, v := range psink.Batch.ValueStrings {
		// (?,?,?)
		inner := strings.Split(string(v), ",")

		for j, v := range inner {
			//(?,
			counter++
			n := strconv.Itoa(counter)

			inner[j] = strings.ReplaceAll(v, "?", "$"+n)
		}
		s = append(s, strings.Join(inner, ","))

	}

	return strings.Join(s, ",")
}

// it autmtically excute the batch if it is full
func (psink *PsqlSink) addKeywordEvent(e KeywordEvent) error {

	psink.Batch.ValueStrings = append(psink.Batch.ValueStrings, "(?,?,?,?,?,?,?,?,?,?,?)")
	psink.Batch.ValueArgs = append(psink.Batch.ValueArgs, e.chainId, e.height, e.txIndex, e.eventIndex, e.attriIndex, e.compositeKey, e.eventType, e.attriKey, e.attriValue, e.txHash, e.keyword)

	if len(psink.Batch.ValueStrings) >= psink.Batch.Size {

		err := psink.ExecBatch()
		if err != nil {

			return fmt.Errorf(">Failed to execute batch insert %v  ", err)

		}

	}

	return nil
}
func (psink *PsqlSink) ExecBatch() error {

	b := psink.Batch

	if len(b.ValueStrings) == 0 {

		return nil
	}

	tx, err := psink.EventSink.DB().Begin()

	if err != nil {

		return err
	}

	stmt := fmt.Sprintf("%s %s", b.Statement, psink.prepareValueStrings())
	_, err = tx.Exec(stmt, b.ValueArgs...)

	if err != nil {
		tx.Rollback()
		fmt.Printf("stmt %s \n b.valueArgs: %v\n", stmt, b.ValueArgs)

		return fmt.Errorf(">Erros in exec batch: %v", err)
	}

	err = tx.Commit()
	//reset batch

	psink.Batch.ValueStrings = nil
	psink.Batch.ValueArgs = nil

	return err

}

// Insert the tx information and entire event
// height is iTh  in a blockstore,
// txIndex is iTh element in deliverTxs of a block,
// eventIndex is iTh event in events of a tx

func (psink *PsqlSink) insertTxEvent(height int64, txIndex int, eventIndex int, txHash string, keyword string, event abcitypes.Event) error {

	if event.Type == "" {
		return nil
	}
	// Add any attributes flagged for indexing.
	for attriIndex, a := range event.Attributes {

		compositeKey := event.Type + "." + a.Key

		// "INSERT INTO keyword_event(chain_id,height,tx_id, event_id,composite_key,type,attribute, value, tx_hash, keyword)"
		// sequence should match with above sql

		chainId := psink.Config.ChainID

		keywordEvent := KeywordEvent{chainId, height, txIndex, eventIndex, attriIndex, compositeKey, event.Type, a.Key, a.Value, txHash, keyword}

		err := psink.addKeywordEvent(keywordEvent)
		if err != nil {
			return err
		}
	}

	return nil
}
func LoadStateAndBlockStore(cfg *config.Config) (*store.BlockStore, state.Store, error) {

	fmt.Printf("Load %s ... ", cfg.DBDir+"blockstore.db")

	dbType := tmdb.BackendType(cfg.DBBackend)
	// load Block Store
	if !tmos.FileExists(cfg.DBDir + "blockstore.db") {
		return nil, nil, fmt.Errorf("no blockstore.db found in %s", cfg.DBDir)
	}

	//data is the directory
	blockStoreDB, err := tmdb.NewDB("blockstore", dbType, cfg.DBDir)

	if err != nil {

		return nil, nil, err
	}

	blockStore := store.NewBlockStore(blockStoreDB)

	fmt.Printf("Done\n")

	// load state store
	fmt.Printf("Load %s ... ", cfg.DBDir+"state.db")

	if !tmos.FileExists(cfg.DBDir + "state.db") {

		return nil, nil, fmt.Errorf("no state.db found in %s", cfg.DBDir)
	}

	//data is the directory
	stateDB, err := tmdb.NewDB("state", dbType, cfg.DBDir)
	if err != nil {
		return nil, nil, err
	}

	stateStore := state.NewStore(stateDB)

	fmt.Printf("Done\n")

	return blockStore, stateStore, nil

}
