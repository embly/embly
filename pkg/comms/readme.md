# Comms

there is a master process

each function has a mailbox and messages can be sent/recieved

we should write an in-process version where our buffers eventually get replaced with
buffered unix sockets

---

## New Syscalls

`_read(id, )`: read from a stream. never blocks, has io:EOF or ErrWouldBlock, similar to rust channel semantics

`_write(id, )`: write, never blocks, always writes the same as before

`_events()`: returns pending events, can be used by `.wait()` calls to wait for a specific id
