# mgolock

mgolock is a package that allows you to add a lock to mongo. When a lock
is acquired, for the duration given by `ttl`, the lock can only be
re-acquired by the same PID on the same host. When ttl is passed, any
other process can acquire it.

To keep a lock once acquired, you must re-acquire it before the `ttl` is
passed.
