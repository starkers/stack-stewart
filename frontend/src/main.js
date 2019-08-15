import Vue from 'vue'
import './plugins/vuetify'
import App from './App.vue'
import router from './router'
import crono from 'vue-crono';

Vue.use(crono);
Vue.config.productionTip = false

new Vue({
  router,
  render: h => h(App)
}).$mount('#app')
