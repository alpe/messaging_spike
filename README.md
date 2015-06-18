# Spike: Async message handling 
Showcase DDD and EIPatterns to deal with some common asynchronous and concurrent update issues.

## Concurrent updates
### Optimistic locking
### Vector clocks
https://en.wikipedia.org/wiki/Vector_clock

## Building state with random message order
* Scenario:
When in sourcing mode, events are received in almost random order. One specific event type leads to an action where others
are required to provide the required data.
Constraints: When we consume an event it is marked as read and not consumed again during processing. In sourcing mode we receive
all events but are not allowed to write events.

## Author
@alpe1
