
// this hack proxies the requets to the server.. no CORS on localhost
module.exports = {
  devServer: {
    proxy: {
      '^/': {
        target: 'http://localhost:8080',
        ws: true,
        changeOrigin: true
      }
    }
  }
}

// USAGE: calling unknown URIs now will proxy this onto the host+port above..
//
// EG.. this "/foo" actually goes to my api on http://localhost:8080/foo now
//
//   methods: {
//     async getDataFoo() {
//       try {
//         const response = await fetch('/foo')
//         const data = await response.json()
//         console.log("hello")
//         this.foo = data
//       } catch (error) {
//         console.error(error)
//       }
//     }
//   }
