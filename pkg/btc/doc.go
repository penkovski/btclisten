// Package btc implements a listening object accepting
// a tcp connection on which it listens for all messages
// broadcasted by connected peer(s).
//
// It performs the initial handshake with the first connected
// peer and then continues to listen for messages on the
// connection.
//
// Currently, connections to other peers are not initiated,
// but it would be nice to ask the initial peer for his peers
// and make multiple other connections and listen for messages
// on all of them. This may make the receiving of data more
// reliable and definitely more interesting for implementation :)
//
// TODO: Create interfaces that would allow message listeners
// to be registered here to receive messages for further
// processing. For example, there may be a database listener
// object that would receive structured messages and save them
// in a relational database, or there may be a listener inserting
// various protocol data extracted from messages in BigQuery.
package btc
