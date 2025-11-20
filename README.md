== A Distributed Auction System ==

::Introduction::

You must implement a **distributed auction system** using replication:
a distributed component which handles auctions, and provides
operations for bidding and querying the state of an auction. The
component must faithfully implement the semantics of the system
described below, and must at least be resilient to one (1) crash
failure.

::MA Learning Goal::

The goal of this mandatory activity is that you learn (by doing) how
to use replication to design a service that is resilient to crashes. In
particular, it is important that you can recognise what the key issues
that may arise are and understand how to deal with them.

 

::API::

Your system must be implemented as some number of nodes, running on
distinct processes (no threads). Clients direct API requests to any
node they happen to know (it is up to you to decide how many nodes can
be known). Nodes must respond to the following API

Method:  bid
Inputs:  amount (an int)
Outputs: ack
Comment: given a bid, returns an outcome among {fail, success or exception}

 

Method:  result
Inputs:  void
Outputs: outcome
Comment:  if the auction is over, it returns the result, else highest bid.

 

::Semantics::

Your component must have the following behaviour, for any reasonable
sequentialisation/interleaving of requests to it:

- The first call to "bid" registers the bidder.

- Bidders can bid several times, but a bid must be higher than the
  previous one(s).

- after a predefined timeframe, the highest bidder ends up as the
  winner of the auction, e.g, after 100 time units from the start of
  the system.

- bidders can query the system in order to know the state of the
  auction.

 

:: Faults :: 

- Assume a network that has reliable, ordered message transport, where
  transmissions to non-failed nodes complete within a known
  time-limit.

- Your component must be resilient to the failure-stop failure of one
  (1) node.