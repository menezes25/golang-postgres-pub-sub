import Vue from 'vue'
import App from './App.vue'
import websocket from "@/services/websocket";
import './assets/css/styles.css'

websocket.startWebSocket();

Vue.config.productionTip = false

new Vue({
  render: h => h(App),
}).$mount('#app')
