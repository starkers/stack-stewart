<template>
  <div class="dashboard">
    <!-- <h1 class="subheading grey--text">Dashboard</h1> -->
    <v-card>
      <v-card-title>
        Deployments
        <v-spacer></v-spacer>

        <v-text-field
          v-model="search"
          append-icon="search"
          label="Search"
          single-line
          hide-details
        ></v-text-field>

      </v-card-title>

      <v-data-table
        :headers="headers"
        :items="stacks"
        :search="search"
        class="elevation-1"
        :rows-per-page-items="[100, 200, 300, 400]"
      >
        <!-- select-all -->


        <template slot="items" slot-scope="props">
          <td>{{ props.item.agent}}</td>
          <td>{{ props.item.name}}</td>
          <td>{{ props.item.namespace}}</td>
          <td>{{ props.item.containers[0].image}}</td>
        </template>


      </v-data-table>
    </v-card>

  </div>
</template>

<script>

// See: https://github.com/iamshaunjp/vuetify-playlist/blob/lesson-16/todo-ninja/src/views/Dashboard.vue#L23
export default {


  data() {
    return {
      search: '',
      itemsPerPage: 100,
      stacks: [],
      headers: [
        // {
        //   text: 'item',
        //   align: 'left',
        //   sortable: false,
        //   value: 'name',
        // },
        { text: 'cluster', value: 'agent' },
        { text: 'name', value: 'name' },
        { text: 'ns', value: 'namespace' },
        { text: 'image', value: 'containers' },
      ],
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
