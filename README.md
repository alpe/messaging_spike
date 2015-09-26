[![Build Status](https://travis-ci.org/alpe/messaging_spike.svg?branch=master)](https://travis-ci.org/alpe/messaging_spike)


# Spike: Async message handling 
Showcase DDD and EIPatterns to deal with some common asynchronous and concurrent update issues.

## Concurrent updates
### Optimistic locking
https://en.wikipedia.org/wiki/Optimistic_concurrency_control
### Vector clocks
https://en.wikipedia.org/wiki/Vector_clock

## Building state with random message order
* Scenario:
When in sourcing mode, events are received in almost random order. One specific event type leads to an action where others
are required to provide the required data.

Constraints: When we consume an event, it is marked as read and not consumed again during processing. In sourcing mode we receive
all events but are not allowed to write events.


## Switch from sourcing to processing mode
* Scenario:
Consumers start in sourcing mode where they consume all previous events they had processed before.
Use case: deployment/ restart to build internal state.

## Message versioning & migration
* Scenario:
Messages are persisted and immutable. When the content structure/ format needs to change all consumers must be able
to handle old and new message formats. 
This "upgrade" logic should be separated (like an anti corruption layer) and not
go into the main processing logic. Otherwise we'd end up with lot of code handling version related conditions across the system. 


## TODO
 - [ ] build state from snapshots
 - [ ] shading consumers
 - [ ] backpressure algo to not overload from a topic (fair consumers)
 - [ ] handy-Lamport algorithm: mini batches
## Resources
* http://basho.com/posts/technical/why-vector-clocks-are-hard/
* https://en.wikipedia.org/wiki/Lamport_timestamps
* https://en.wikipedia.org/wiki/Vector_clock
 
## Author
@alpe1
