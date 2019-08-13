<template>
  <div class="dashboard">
    <h1 class="subheading grey--text">Dashboard</h1>

    <v-container class="my-5">

      <v-card flat v-for="stack in stacks" :key="stack.trace">
        <v-layout row wrap class="pa-3 state complete">
          <v-flex xs12 md4>
            <div class="caption grey--text">trace</div>
            <div>{{ stack.trace }}</div>
          </v-flex>
          <v-flex xs6 sm4 md2>
            <div class="caption grey--text">lane</div>
            <div>{{ stack.lane }}</div>
          </v-flex>
          <v-flex xs6 sm4 md2>
            <div class="caption grey--text">name</div>
            <div>{{ stack.name }}</div>
          </v-flex>
          <v-flex xs6 sm4 md2>
            <div class="caption grey--text">namespace</div>
            <div>{{ stack.namespace }}</div>
          </v-flex>
          <v-flex xs2 sm4 md2>
            <div class="right">
              <v-chip small class="updating white--text my-2 caption">
                {{stack.replicas.available}} / {{stack.replicas.ready}} / {{stack.replicas.updated}}
              </v-chip>
            </div>
          </v-flex>
        </v-layout>
        <v-divider></v-divider>
      </v-card>

    </v-container>

  </div>
</template>

<script>

// See: https://github.com/iamshaunjp/vuetify-playlist/blob/lesson-16/todo-ninja/src/views/Dashboard.vue#L23
export default {


  data() {
    return {
      stacks: []
    }
  },
  mounted() {
    this.getStacks()
  },
  methods: {
    async getStacks() {
      try {
        const response = await fetch('/stacks')
        const data = await response.json()
        console.log("hello world")
        this.stacks = data
      } catch (error) {
        console.error(error)
      }
    }
  }
}
</script>

<style>

.state.complete{
  border-left: 4px solid #3cd1c2;
}
.state.updating{
  border-left: 4px solid #ffaa2c;
}
.state.errored{
  border-left: 4px solid #f83e70;
}
.v-chip.complete{
  background: #3cd1c2;
}
.v-chip.updating{
  background: #ffaa2c;
}
.v-chip.errored{
  background: #f83e70;
}

</style>
