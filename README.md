# Query Tracer

QueryTracer is a simple MySQL proxy written in Go that listens for, and logs queries on a per-table basis.

WARNING: SSL does not currently work with this proxy. The proxy explicitly modifies the server capabilities packet to remove SSL support. PR's adding SSL support is welcome.
