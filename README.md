
[![Build Status](https://travis-ci.org/starkers/stack-stewart.svg?branch=master)](https://travis-ci.org/starkers/stack-stewart)

```
├── cmd
│   ├── agent  <--- kubernetes agent  (runs as deployment, sends data to server)
│   ├── mock   <--  just a script that can send test data to the server (currently)
│   └── server <--- http API (receives agent data + serves static assets)
└── shared     <-   golang shared structs
```
