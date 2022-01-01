# Example from
https://coussej.github.io/2015/09/15/Listening-to-generic-JSON-notifications-from-PostgreSQL-in-Go/

psql: > notify "events", '{"foo" : "bar"}';
or    > notify events, '{"foo": "bar"}';
or
$ usql 'postgresql://webapp:webapp@localhost/webapp?sslmode=disable' -c "notify events, '{\"foo\": \"bar bar 2\"}';"


## Changes:
 * When PG is restarted a nil notification is received from the channel --> added handling of that
 * Q: Why was the Ping() call in a separate goroutine?
 * Experimented with responding to failed Ping() with reconnecing. This seems unnecessary.
 * Added a mechanism for shutting down the listener via `shutdownChan`.

 2nd round:
 * added http server with websocket upgrade

 3rd round:
 * Remodelled after https://github.com/gorilla/websocket/blob/master/examples/chat
   This seems is the ultimate "howto" for gorilla websockets.
