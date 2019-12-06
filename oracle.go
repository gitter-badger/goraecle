package goraecle

import (
	"log"
	"os"
	"strings"

	"github.com/aeternity/aepp-sdk-go/v7/account"
	"github.com/aeternity/aepp-sdk-go/v7/aeternity"
	"github.com/aeternity/aepp-sdk-go/v7/binary"
	"github.com/aeternity/aepp-sdk-go/v7/config"
	"github.com/aeternity/aepp-sdk-go/v7/naet"
	"github.com/aeternity/aepp-sdk-go/v7/swagguard/node/models"
	"github.com/aeternity/aepp-sdk-go/v7/transactions"
	"github.com/pkg/errors"
)

var (
	// DebugMode enables debug logging for the defaultOracle
	DebugMode bool

	// AeDebugging till turn on node debugging when set to true
	AeDebugging = false
)

// Oracle defines parameters for running an oracle.
type Oracle struct {
	// OraclePubKey is the oracles's public key.
	OraclePubKey string

	// Handler is the handler to invoke on query requests.
	// DefaultOracleMux is used when nil.
	Handler Handler

	// QueryParser is the raw query parser to invoke.
	// DefaultQueryParser is used when nil.
	QueryParser QueryParser

	Debug  bool
	Logger *log.Logger

	acc       *account.Account
	networkID string
	node      *naet.Node // TODO: Change to node interface
	aeOracle  *models.RegisteredOracle
	//TODO: doneChan, onShutdown
	//TODO: panic recovery
}

// AeternityNetwork is a custom type representing the aeternity network.
type AeternityNetwork string

// String implements the stringer interface and returns the aeternity network id
// of the AeternityNetwork value.
func (n *AeternityNetwork) String() string {
	return string(*n)
}

// NodeURL returns the defauly node url for the AeternityNetwork value.
func (n AeternityNetwork) NodeURL() string {
	switch n {
	case AeternityMainnet:
		return "http://sdk.aepps.com"
	case AeternityTestnet:
		return "http://sdk-testnet.aepps.com"
	case AeternityLocalnet:
		return "http://localhost:3015"
	}
	return ""
}

const (
	// AeternityMainnet represents the aeternity mainnet
	AeternityMainnet = AeternityNetwork("ae_mainnet")

	// AeternityTestnet represents the aeternity testnet
	AeternityTestnet = AeternityNetwork("ae_uat")

	// AeternityLocalnet represents an aeternity localnet
	// TODO: Proper nodeURL overriding
	AeternityLocalnet = AeternityNetwork("ae_localnet")
)

// NewOracle returns an initialized Oracle.
// TODO: options
func NewOracle(privateKey string, network AeternityNetwork) (*Oracle, error) {
	acc, err := account.FromHexString(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "account from hex string")
	}

	o := Oracle{
		acc:          acc,
		OraclePubKey: strings.Replace(acc.Address, "ak_", "ok_", 1),
	}

	o.Network(network)
	return &o, nil
}

// Network changes the network to use for the Oracle.
func (o *Oracle) Network(n AeternityNetwork) {
	config.Node.NetworkID = n.String()
	o.networkID = n.String()
	o.node = naet.NewNode(n.NodeURL(), AeDebugging)
}

// ttlNoncer returns a generated TTLNoncer for the node
func (o *Oracle) ttlNoncer() transactions.TTLNoncer {
	_, _, ttlnoncer := transactions.GenerateTTLNoncer(o.node)
	return ttlnoncer
}

// RegisterExisting queries the node for the oracle by it's public key.
func (o *Oracle) RegisterExisting() error {
	oracle, err := o.node.GetOracleByPubkey(o.OraclePubKey)
	if err != nil {
		return errors.Wrap(err, "getting registered oracle")
	}
	o.aeOracle = oracle

	return nil
}

// Register checks whether the oracle has already been registered.
// If the oracle isn't already registered Register will take care of registration of a new oracle.
func (o *Oracle) Register() error {
	// Check if already registerd.
	if err := o.RegisterExisting(); err == nil {
		// Already registered, so we're done.
		return nil
	}

	// Not registered, so register as a new Oracle.
	// TODO: QuerySpec, ResponseSpec
	reg, err := transactions.NewOracleRegisterTx(o.acc.Address, "hello", "helloback", config.Client.Oracles.QueryFee, config.OracleTTLTypeDelta, config.Client.Oracles.OracleTTLValue, config.Client.Oracles.ABIVersion, o.ttlNoncer())
	if err != nil {
		return errors.Wrap(err, "register new oracle")
	}

	// Sign the registration transaction.
	_, _, _, _, _, err = aeternity.SignBroadcastWaitTransaction(reg, o.acc, o.node, o.networkID, config.Client.WaitBlocks)
	if err != nil {
		return errors.Wrap(err, "signing transaction")
	}

	// Get and set the registration.
	o.RegisterExisting()

	return nil
}

func (o *Oracle) debug(v interface{}) {
	if o.Debug && o.Logger != nil {
		o.Logger.Println(v)
	}
}

func (o *Oracle) debugf(format string, v ...interface{}) {
	if o.Debug && o.Logger != nil {
		o.Logger.Printf(format, v...)
	}
}

func (o *Oracle) processQuery(q string) (string, error) {
	// IF a custom handler has been set use that one,
	// else use the default handler.
	handler := o.Handler
	if handler == nil {
		handler = DefaultOracleMux
	}

	// If a custom query parser has been set use that one,
	// else use DefaultQueryParser
	queryParser := o.QueryParser
	if queryParser == nil {
		queryParser = DefaultQueryParser
	}

	// Decode the query
	b, err := binary.Decode(q)
	if err != nil {
		return "", errors.Wrap(err, "decoding")
	}
	o.debugf("received query: %s", b)

	// Send the query to the raw query handler, which will return a formatted request.
	r, err := queryParser.ParseOracleQuery(string(b)) // TODO: Error handling
	if err != nil {
		return "", errors.Wrap(err, "RawQueryHandler")
	}

	// Send the request to the handler for an answer or error
	answ, err := handler.ServeOracle(r)
	if err != nil {
		return "", errors.Wrap(err, "ServeOracle")
	}

	return answ, nil
}

// ListenAndServe listens for incoming oracle queries and then calls the RawQueryHandler to handle incoming requests.
func (o *Oracle) ListenAndServe() error {
	// Register the oracle if no oracle is set.
	if o.aeOracle == nil {
		o.Register()
	}
	o.debugf("invoked ListenAndServe, oracle: %+v\n\n", o)

	var readUntilPosition int // TODO: Position storage
	// o.debugf("Initialized. handler: %T, rawQueryHandler: %T, readUntilPosition: %d", handler, rawQueryHandler, readUntilPosition)

	// Wait for incoming requests.
	o.debugf("oracle pub key: %v", o.OraclePubKey)
	for {
		// Get queries from the node.
		oQueries, err := o.node.GetOracleQueriesByPubkey(o.OraclePubKey)
		if err != nil {
			o.debugf("oQueries error: %v", err)
			continue // TODO: Error handling
		}

		if len(oQueries.OracleQueries) > 0 {
			o.debugf("oQueries: %+v", oQueries)
		}

		//  Loop over each query
		for _, q := range oQueries.OracleQueries[readUntilPosition:] {
			// Process query
			answ, err := o.processQuery(*q.Query)
			if err != nil {
				o.debugf("binary.Decode error: %v", err)
				continue // TODO: Error handling
			}

			// Send the answer as a response.
			response, err := transactions.NewOracleRespondTx(o.acc.Address, o.OraclePubKey, *q.ID, answ, config.OracleTTLTypeDelta, config.Client.Oracles.ResponseTTLValue, o.ttlNoncer())
			if err != nil {
				o.debugf("transactions.NewOracleRespondTx error: %v", err)
				continue // TODO: Error handling
			}

			// Sign the response transaction.
			_, _, _, _, _, err = aeternity.SignBroadcastWaitTransaction(response, o.acc, o.node, o.networkID, config.Client.WaitBlocks)
			if err != nil {
				o.debugf("aeternity.SignBroadcastWaitTransaction error: %v", err)
				continue // TODO: Error handling
			}

			readUntilPosition++
		}
	}
}

// TestQuery is a temporary function for sending a query request
// TODO: REMOVE
func (o *Oracle) TestQuery() (string, error) {
	query, err := transactions.NewOracleQueryTx(o.acc.Address, o.OraclePubKey, "hello:name=Arjan van Eersel", config.Client.Oracles.QueryFee, 0, 100, 0, 100, o.ttlNoncer())
	if err != nil {
		return "", err
	}

	_, _, _, _, _, err = aeternity.SignBroadcastWaitTransaction(query, o.acc, o.node, o.networkID, config.Client.WaitBlocks)
	if err != nil {
		return "", err
	}

	return "", nil
}

// GetAnswer is a temporary function to get all answers to a query.
// TODO: REMOVE
func (o *Oracle) GetAnswer(queryID string) (string, error) {
	oQueries, err := o.node.GetOracleQueriesByPubkey(o.OraclePubKey)
	if err != nil {
		return "", err
	}

	for _, q := range oQueries.OracleQueries {
		if q.Response != nil {
			answ, err := binary.Decode(*q.Response)
			if err != nil {
				return "", err
			}
			return string(answ), nil
		}
	}

	return "", nil
}

// defaultOracle is the default oracle instance provided.
var defaultOracle *Oracle

// ListenAndServe starts the defaultOracle to listen for query requests.
func ListenAndServe(privateKey string, network AeternityNetwork) error {
	o, err := NewOracle(privateKey, network)
	if err != nil {
		return err
	}

	if DebugMode {
		o.Debug = true
		o.Logger = log.New(os.Stdout, "defaultOracle : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	}

	defaultOracle = o

	return defaultOracle.ListenAndServe()
}
