<template>
  <v-app dark class="grey darken-1">
    <Navbar />


    <v-content class="mx-3 mb-4">
      <router-view></router-view>
    </v-content>

  </v-app>
</template>

<script>
import Navbar from './components/Navbar'

export default {
  components: { Navbar },
  name: 'App',
  data () {
    return {
    }
  }
}
</script>
