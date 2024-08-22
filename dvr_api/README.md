<h1>dvr_api</h1>

https://github.com/valyala/fastjson <-- maybe use this for speed

application that listens on a tcp server for messages from devices, a websocket server for messages from websocket clients, and on an http server for requests for data from whoever wants it

device server pipes messages into a database

websocket server pipes messages to devices connected to the device server

api server responds to requests for data over HTTP get

<h1>websocket pub/sub to device messages feature</h1>

will have to unmarshal json into structs

make all messages have to be in json

subscribe by sending a WS message like { subscribe: {"dev_id1", "dev_id2"}}, empty to remove subscription
