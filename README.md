


```
├── cmd
│   ├── agent  <--- kubernetes agent  (runs as deployment, sends data to server)
│   ├── mock   <--  just a script that can send test data to the server (currently)
│   └── server <--- http API (receives agent data + serves static assets)
└── shared     <-   golang shared structs
```
